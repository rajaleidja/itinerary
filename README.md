# Itenerary iteneraryWithBonuses.go

***

## 1 task Golang
### Nikita Strekalov

___
> "Every flight needs a good and accurate schedule. I hope this schedule was well edited."
>
> Nikita Strekalov
___

## Help Message

If no argument is provided or argument "help", the tool displays a brief help message explaining how to use the application:
```bash
$ go run iteneraryWithBonuses.go -h
Usage: ./input.txt ./output.txt ./airport-lookup.csv
```
Also try "-b -h". There are some interesting things
```bash
$ go run iteneraryWithBonuses.go -h -b
```

## How to use with bonus

```bash
go run iteneraryWithBonuses.go -b ./input.txt ./output.txt ./airport-lookup.csv
```
then input column numbers

# IMPORTANT
there are 2 files.
file "iteneraryWithBonuses.go" is done with AND without bonuses. Depends from "-b" flag
file main.go is just task without bonuses.

"iteneraryWithBonuses.go":
- [X] It has other non-specific interesting bonuses
- color formatting
- bold serif in some places
- [X] It converts city names from airport codes
- *# and *##
- [X] It works with non-standard airport lookup column order
- Need to input column numbers
- [X] It makes good use of formatting
- outputting into terminal with colored text
- [X] can use with and without bonus task
- Use go run . -b -h to see more. EXAMPLE OF USAGE WITH BONUS TASKS: go run . -b ./input.txt ./output.txt ./airport-lookup.csv