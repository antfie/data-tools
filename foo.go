package main

// this was to fix up filenames where we did not have ff/ff type prefixes

//
//func (ctx *Context) Foo() {
//	//for {
//	//		var hashes []string
//	//		result := ctx.DB.Raw(`
//	//SELECT		fh.hash
//	//FROM 		file_hashes fh
//	//WHERE		fh.zapped = 1
//	//AND			fh.ignored = 0
//	//ORDER BY	fh.id -- for deterministic result order
//	//`).Scan(&hashes)
//	//
//	//		if result.Error != nil {
//	//			log.Fatal(result.Error)
//	//		}
//	//
//	//		if hashes == nil {
//	//			return
//	//		}
//	//
//	//		if len(hashes) == 0 {
//	//			return
//	//		}
//
//	hashes, err := os.ReadDir("/Volumes/OLD_Backup/ZAP/")
//
//	if err != nil {
//		return
//	}
//
//	bar := progressbar.Default(int64(len(hashes)))
//
//	//orchestrator := utils.NewTaskOrchestrator(bar, len(hashes), ctx.Config.MaxConcurrentFileOperations)
//
//	for _, hash := range hashes {
//		//orchestrator.StartTask()
//		//go func() {
//		hexFileName := hash.Name()
//		sourcePath := path.Join("/Volumes/OLD_Backup/ZAP/", hexFileName)
//		_, err := os.Stat(sourcePath)
//
//		if os.IsNotExist(err) {
//			err = bar.Add(1)
//
//			if err != nil {
//				log.Printf("failed to update progress bar: %v", err)
//			}
//			//orchestrator.FinishTask()
//			continue
//		}
//
//		if err != nil {
//			log.Print(err)
//			continue
//		}
//
//		ty, err := GetFileType(sourcePath)
//
//		if err != nil {
//			log.Print(err)
//			continue
//		}
//
//		print(ty)
//
//		destinationPath := path.Join("/Volumes/OLD_Backup/ZAP2/", hexFileName[:2], hexFileName[2:4], hexFileName[4:])
//
//		err = CopyOrMoveFile(sourcePath, destinationPath, true)
//
//		if err != nil {
//			log.Fatalf("Could not Foo file \"%s\": %v", sourcePath, err)
//		}
//
//		err = bar.Add(1)
//
//		if err != nil {
//			log.Printf("failed to update progress bar: %v", err)
//		}
//
//		//orchestrator.FinishTask()
//		//}()
//	}
//
//	//orchestrator.WaitForTasks()
//	//}
//}
