IF EXIST panic.html (
    del panic.html
)
ankathost 2>&1 | pp --html=panic.html
IF EXIST panic.html (
    start panic.html
)