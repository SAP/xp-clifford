Output example.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 600

Type "go run main.go export" Sleep 500ms Enter
Sleep 4s

# Select option-3 and option-5 without selecting their parents.
Down Sleep 500ms
Down Sleep 500ms
Space Sleep 1s
Down Sleep 500ms
Down Sleep 500ms
Space Sleep 2s

# Filter for option-4, finish editing, and select it.
Type "/" Sleep 500ms
Type "option-4" Sleep 2s
Enter Sleep 500ms
Space Sleep 2s

# Clear the filter to show all three selected leaves, then confirm.
Escape Sleep 3s
Enter
Sleep 4s
