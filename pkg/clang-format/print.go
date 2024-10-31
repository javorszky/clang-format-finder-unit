package clang_format

//
// func Print() error {
// 	resolvedOptions := generateCombinations(options)
//
// 	fmt.Printf("resolvedoptions: %v\n", resolvedOptions)
//
// 	// delete all files and directories in the target directory
// 	err := deleteContents(TargetDirectory)
// 	if err != nil {
// 		return errors.Wrap(err, "deleteContents")
// 	}
//
// 	_, err = os.Stat(TargetDirectory)
// 	if os.IsNotExist(err) {
// 		err = os.Mkdir(TargetDirectory, 0755)
// 		if err != nil {
// 			return errors.Wrap(err, "os.Mkdir targetDirectory")
// 		}
// 	}
//
// 	for i, o := range resolvedOptions {
// 		fmt.Printf("Trying to write file for %04d\n", i)
// 		err = os.WriteFile(
// 			path.Join(TargetDirectory, fmt.Sprintf(filename, fmt.Sprintf("%04d", i))),
// 			[]byte(o.String()),
// 			0755,
// 		)
// 		if err != nil {
// 			return errors.Wrapf(err, "os.WriteFile %04d", i)
// 		}
// 	}
//
// 	return nil
// }
