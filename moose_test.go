package main

//
//func TestMooose(t *testing.T) {
//	//for x:= range( fff())
//	//for lx := range GenerateHashBaseDirs {
//	//	print(lx)
//	//}
//
//	sd := "003a081f9c79b45830fca525d2be0878576cbf0c3e1abbbeb3c84d0a32cb5afe90f25ea031cfc9eeb2c7d332af58abe093b94fd28247ca88c25d75c71791572d"
//	xll := FormatRelativeZapFilePathFromHash(sd)
//	print(xll)
//	for nn := range GetAllZappedFilesImproved("/Volumes/Backup/ZAP") {
//		print(nn)
//	}
//}
//
//func GetAllZappedFiles(zapPath string) iter.Seq2[string, error] {
//	return func(yield func(string, error) bool) {
//		for prefix := range GenerateZapHashDirectoryPrefixes {
//			basePath := path.Join(zapPath, prefix)
//			entries, err := os.ReadDir(basePath)
//
//			if err != nil {
//				yield("", err)
//				return
//			}
//
//			for _, e := range entries {
//				filePath := path.Join(basePath, e.Name())
//
//				if e.IsDir() {
//					yield("", errors.New(fmt.Sprintf("a subdirectory was unexpected \"%s\" in the Zap path", filePath)))
//					return
//				}
//
//				yield(filePath, nil)
//			}
//		}
//	}
//}
