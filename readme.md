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
from unit's current master version (https://github.com/nginx/unit/commit/1e345b3477ef8ca2eb79a4384dda1858a5a84e41)
as of 5th November 2024 with the given options.

There's also a file called [.clang-format-doesntmatter](.clang-format-doesntmatter) which lists all the options 
where the different values for those options did not change the number of lines changed. It does not mean that those 
options ended up with the _same_ diff, it means they changed the same number of lines.

## FAQ

### How does this even work?

For every option, it picks one value, generates a `.clang-format` file, runs the formatter, gets the diff, and 
counts the lines changed across all files, then picks the value with the lowest number of lines changed for a given 
option before it moves to the next option.

Before it checks each value for each option it git resets the repository to the current HEAD of the master branch, 
to make sure there are no changes from a previous run.

It goes through the entire option list three times, to make sure that the file we arrive on is the best 
constellation of options / values we can have.

### How are lines changed calculated?

After running the tool we grab the diff with [`git diff --numstat`](https://git-scm.com/docs/git-diff#Documentation/git-diff.txt---numstat) which we parse. For each file there's a pair of numbers: added and deleted. I take the higher of these with the assumption that if we added 5 lines and deleted 4 lines, we actually only changed 5 lines (changed 4, added 1).