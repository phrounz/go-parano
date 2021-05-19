package src

import (
	"fmt"
	"os"
	"path/filepath"

	"./fileparser"
	"./util"
)

//------------------------------------------------------------------------------

type packageInfos struct {
	packageName      string
	packageDir       string
	subPackagesInfos []*packageInfos
	infosByFile      map[string]infosFile
}

type infosFile struct {
	packageName              string
	rootNode                 *fileparser.Node
	fileConstants            []fileparser.ConstValue
	featurePrivateToFile     *featurePrivateToFile
	featureExhaustiveFilling *featureExhaustiveFilling
}

//------------------------------------------------------------------------------

func DoAll(pkgDir string, ignoreGoFiles map[string]bool, sqlqo SQLQueryOptions) {

	var rootPkg = recurseDir(pkgDir, ignoreGoFiles, sqlqo)

	ParanoSqllintCheckQueries(sqlqo)

	var mInfosByPackageName = make(map[string]*packageInfos)
	processPkgRecursiveAndMakeMap(rootPkg, mInfosByPackageName)

	if util.IsInfo() {
		util.Info("\"Fourth\" pass")
	}

	processPkgAgain(mInfosByPackageName)
}

//------------------------------------------------------------------------------

func processPkgRecursiveAndMakeMap(pkgInfos *packageInfos, mInfosByPackageName map[string]*packageInfos) {
	mInfosByPackageName[pkgInfos.packageName] = pkgInfos
	for _, subPackageInfos := range pkgInfos.subPackagesInfos {
		processPkgRecursiveAndMakeMap(subPackageInfos, mInfosByPackageName)
	}
}

//------------------------------------------------------------------------------

func recurseDir(pkgDir string, ignoreGoFiles map[string]bool, sqlqo SQLQueryOptions) *packageInfos {

	var subPackagesInfos = make([]*packageInfos, 0)

	var items, err = filepath.Glob(pkgDir + "/*")
	if err != nil {
		panic(err)
	}
	var srcFiles []string
	for _, item := range items {
		if len(item) >= 3 && item[len(item)-3:] == ".go" {
			srcFiles = append(srcFiles, item)
		} else {
			info, err := os.Stat(item)
			if os.IsNotExist(err) {
				panic("File does not exist: " + item)
			}
			if info.IsDir() {
				subPackagesInfos = append(subPackagesInfos, recurseDir(item, ignoreGoFiles, sqlqo))
			}
		}
	}
	// if len(srcFiles) == 0 {
	// 	fmt.Println("WARNING: no sources files in " + pkgDir)
	// }
	if util.IsInfo() {
		util.Info("Processing package: %s", pkgDir)
	}

	var infosByFile = processPkgFiles(srcFiles, ignoreGoFiles, sqlqo)
	var packageName string // note: remains empty string if no source files
	for _, infosFile := range infosByFile {
		packageName = infosFile.packageName
		break
	}
	return &packageInfos{
		packageName:      packageName,
		packageDir:       pkgDir,
		subPackagesInfos: subPackagesInfos,
		infosByFile:      infosByFile,
	}
}

//------------------------------------------------------------------------------

func processPkgFiles(files []string, ignoreGoFiles map[string]bool, sqlqo SQLQueryOptions) (infosByFile map[string]infosFile) {

	infosByFile = make(map[string]infosFile)
	for _, filename := range files {
		if _, ok := ignoreGoFiles[filename]; !ok {
			infosByFile[filename] = processFile(filename)
		}
	}

	if util.IsDebug() || util.IsInfo() {
		util.Info("  Checking ...")
	}

	//----
	// third pass => check

	var failedAtLeastOnce = false
	for filename1, fileInfos := range infosByFile { // for each input file
		fileInfos.rootNode.Visit(func(n *fileparser.Node) {
			if len(sqlqo.FunctionsNames) > 0 {
				failedAtLeastOnce = ParanoSqllintVisit(n, filename1, fileInfos.fileConstants, sqlqo) && failedAtLeastOnce
			}
			if n.Name != "" {
				for filename2, fileInfos2 := range infosByFile { // for each file
					failedAtLeastOnce = ParanoPrivateToFileCheck(n, fileInfos2.featurePrivateToFile, filename1, filename2) && failedAtLeastOnce
					failedAtLeastOnce = ParanoExhaustiveFillingCheck(n, fileInfos2.packageName, fileInfos2.featureExhaustiveFilling, filename1, filename2) && failedAtLeastOnce
				}
			}
		})
	}
	if failedAtLeastOnce {
		fmt.Println("Stopping program now.")
		os.Exit(1)
	}

	return
}

//------------------------------------------------------------------------------

func processFile(filename string) infosFile {
	if util.IsDebug() {
		util.DebugPrintf("===============================> %s", filepath.Base(filename))
	}
	if util.IsInfo() {
		util.Info("  Scanning: " + filepath.Base(filename) + " ...")
	}

	//----
	// first pass => load file and get informations

	var fileInfo = fileparser.ReadFile(filename)

	//----
	// init

	var featurePrivateToFile = ParanoPrivateToFileInit(fileInfo.FileBuffer)
	var featureExhaustiveFilling = ParanoExhaustiveFillingInit()

	//----
	// second pass => gather informations about nodes of this file

	fileInfo.RootNode.Visit(func(n *fileparser.Node) {
		if util.IsDebug() {
			n.Display()
		}
		ParanoPrivateToFileVisit(n, featurePrivateToFile)
		ParanoExhaustiveFillingVisit(n, featureExhaustiveFilling)
	})

	var infosf = infosFile{
		packageName:              fileInfo.PackageName,
		rootNode:                 fileInfo.RootNode,
		fileConstants:            fileInfo.FileConstants,
		featurePrivateToFile:     featurePrivateToFile,
		featureExhaustiveFilling: featureExhaustiveFilling,
	}

	if util.IsDebug() {
		util.DebugPrintf("\n\n\n%+v\n", infosf)
	}

	return infosf
}

//------------------------------------------------------------------------------

func processPkgAgain(mInfosByPackageName map[string]*packageInfos) {

	//----
	// fourth pass => check

	ParanoExhaustiveFillingCheckGlobal(mInfosByPackageName)
}

//------------------------------------------------------------------------------
