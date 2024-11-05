# Go command line tool to churn clang-format results

How to use:

1. clone nginx/unit into the `unit` directory. It's in the gitignore file and is assumed to be there with `git clone 
git@github.com:nginx/unit.git` from the same directory this readme file is in
2. generate the `files.list` file with the following command: `find unit/src -type f \( -name "*.c" -o -name "*.h" \) > files.list`
3. run `make run` and let it churn on the code, it will check options in three passes, and then checks a number of 
   additional options to get to a file that changes the lowest number of lines
4. get the results back

## The results

You can find the ideal clang-format file in [.clang-format-ideal](.clang-format-ideal). It changes **17,666** lines 
with the given options.

There's also a file called [.clang-format-doesntmatter](.clang-format-doesntmatter) which lists all the options 
where the different values for those options did not change the number of lines changed. It does not mean that those 
options ended up with the _same_ diff, it means they changed the same number of lines.