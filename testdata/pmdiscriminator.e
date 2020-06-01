# Test P-M discriminator steppers alluded to in Clippinger '48 "A Logical
# Coding System Applied to the ENIAC".
#
# It's unclear how many of these existed, but the report mentions "no. 2" so
# in theory there were at least two.
#
# Expect a18=1 (-42 is negative)
#    and a19=1 (+42 is positive)

p pm1.i  A-1
p pm1.di D-1
p pm1.1o B-1  # positive
p pm1.2o B-2  # negative

p pm2.cdi A-2
p pm2.i   A-3
p pm2.di  D-1
p pm2.1o  B-3  # positive
p pm2.2o  B-4  # negative

# Transmit a1
p a1.A 1
p A-1 a1.5i
s a1.op5 A
p a1.5o A-2
p A-2 a1.6i
s a1.op6 0
s a1.cc6 0
s a1.rp6 1
p a1.6o A-3
# Then transmit a2
p a2.A 1
p A-3 a2.5i
s a2.op5 A
# Select PM digit
p 1 ad.dp.1
p ad.dp.1.11 D-1

p B-1 a17.1i
s a17.op1 ε
s a17.cc1 C
p B-2 a18.1i
s a18.op1 ε
s a18.cc1 C
p B-3 a19.1i
s a19.op1 ε
s a19.cc1 C
p B-4 a20.1i
s a20.op1 ε
s a20.cc1 C

set a1 -42
set a2 42
p i.io A-1

b i
