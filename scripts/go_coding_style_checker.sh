#!/bin/bash

# Check if gofmt is installed in ./backend
if ! command -v gofmt > /dev/null 2>&1
then
    echo -e "\033[1;31mgofmt could not be found. Please install Go to use this script.\033[0m"
    echo -e "You can install it by running: \033[1;33mgo install golang.org/x/tools/cmd/gofmt@latest\033[0m"
    exit 1
fi

# Check if gofmt is properly formatted (run on repository root)
if ! gofmt -l . | grep -q .
then
    echo -e "\033[1;29mgofmt: All files are properly formatted.\033[0m"
else
    echo -e "\033[1;29mgofmt: The following files are not properly formatted:\033[0m"
    gofmt -l . > "$PWD/scripts/gofmt_output.txt"
    for file in $(cat "$PWD/scripts/gofmt_output.txt"); do
        echo -e "\033[0;31m$file\033[0m"
    done
    exit 1
fi

echo -e "\033[1;32mAll checks passed successfully!\033[0m"
exit 0
