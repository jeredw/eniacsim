# Test the function table selector as described in Clippinger '48 "A Logical
# Coding System Applied to the ENIAC"
#
# This example selects digit 3 of the PC from a6.

p sft.i 1-1
p sft.di 1
p sft.1o 2-9
p sft.2o 2-1
p sft.3o 2-9
p sft.4o 2-9
p sft.5o 2-9
p sft.6o 2-9

# select third digit of a6 on 1
p a6.A ad.permute.1
p ad.permute.1 1
s ad.permute.1 0,0,0,0,0,0,0,0,0,0,3
p 1-1 a6.1i
s a6.op1 A

# (digit selected in a2)
p a2.α 1
p 1-1 a2.1i
s a2.op1 α

# 2-9 means the ft selector decode was wrong.
p a19.α 2
p 2-9 a19.1i
s a19.op1 α
s a19.cc1 C

# 2-1 means the ft selector decoded properly
p a20.α 2
p 2-1 a20.1i
s a20.op1 α
s a20.cc1 C

set a6 142

p i.io 1-1
b i
