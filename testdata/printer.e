# Tests custom printer plugboard wiring
#
# By default plugboard wiring is controlled implicitly by pr P/C settings.  Any
# printing group is wired through to the corresponding columns, and
# non-printing groups are not wired through.  Signs are wired to column 1 of
# the first column in their coupled group.
#
# Instead, custom plugboard wiring may be specified with additional switches.
# Setting any plugboard switch will override any implicit settings based on P/C
# switches, and any unset plugboard switch is assumed to be "nc".

set a13 -9999900000

s pr.2 P

# pr.pm1-80 correspond to plugboard connections that control what punch magnets
# output for each output column.
s pr.pm1 nc   # column 1: not connected, don't print
s pr.pm2 0    # column 2: print zero punch (valid values are 0-12)
# NB there is only room to wire a few fixed digits on the plugboard, but this
# isn't simulated

s pr.pm3 2,1    # column 3: group 2, digit 1 (groups 1-16, digits 1-5)
s pr.pm4 2,2,m2 # column 4: group 2, digit 2 including minus indication for group 1
s pr.pm5 2,3    # column 5: group 2, digit 3
s pr.pm6 2,4    # column 5: group 2, digit 4
s pr.pm7 2,5    # column 5: group 2, digit 5

# should print _00-001 in first 6 columns
p i.io i.pi
b i
