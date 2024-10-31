package clang_format

//
// func generateCombinations(options map[string][]string) []clangFormat {
// 	if len(options) == 0 {
// 		fmt.Printf("returning empty, reached the deep")
// 		return []clangFormat{{}}
// 	}
//
// 	fmt.Printf("options is not null len\n%v\n", options)
//
// 	firstOption := options[0]
// 	fmt.Printf("first option: %v\n", firstOption)
// 	restCombinations := generateCombinations(options[1:])
// 	fmt.Printf("rest combinations: %v\n", restCombinations)
//
// 	var result []clangFormat
// 	fmt.Printf("created empty result slice\n")
// 	for key, values := range firstOption {
// 		for _, value := range values {
// 			for _, combination := range restCombinations {
// 				newCombination := make(clangFormat)
// 				for k, v := range combination {
// 					newCombination[k] = v
// 				}
// 				newCombination[key] = value
// 				fmt.Printf("adding a new thing to the result slice\n")
// 				result = append(result, newCombination)
// 			}
// 		}
// 	}
//
// 	return result
// }

func generateBasic(options map[string][]string) ClangFormat {
	result := make(ClangFormat)

	for key, values := range options {
		result[key] = values[0]
	}

	return result
}
