---
applyTo: '**/*.go'
---
* Respect standard Go formatting and conventions;
* Use `gofmt` to format the code;
* Use `go vet` to check for potential issues in the code;
* When a function call returns an error, prefer the `if err := function(); err != nil {}` pattern over having the call and check on separate lines;
* Group related var and const definitions under the same `var (...)` or `const (...)` block;
* When creating commands with Cobra:
    * don't define a long description unless it's strictly necessary to the comprehension of the command;
    * implement command aliases when it makes sense
