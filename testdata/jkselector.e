# The jk selectors are two notional extra six-stage steppers postulated for
# decoding Sh and Sh' instruction operands in the 60-order code described in
# Clippinger '48 "A Logical Coding System Applied to the ENIAC".
#
# This test decodes "01" and then "95".  a19 and a20 should both be 1.
# A-1  decode (-> B-1 or B-2)
# A-2  B-1 or B-2 propagate through sjk -> C-1 or C-2
# A-4  clear. after A-3 because clear is async
# A-5  decode

p pm1.di D-2
p pm1.i A-10
p pm1.1o B-1  # positive (aka 0)
p pm1.2o B-2  # negative (aka 9)
p pm1.cdi A-4

p sjk1.di D-1
p sjk1.i B-1
p sjk1.1o B-9  # 0
p sjk1.2o C-1  # 1
p sjk1.3o B-9  # 2
p sjk1.4o B-9  # 3
p sjk1.5o B-9  # 4
p sjk1.6o B-9  # 5
p sjk1.cdi A-4

p sjk2.i B-2
p sjk2.di D-1
p sjk2.1o B-8  # 0
p sjk2.2o B-8  # 1
p sjk2.3o B-8  # 2
p sjk2.4o B-8  # 3
p sjk2.5o B-8  # 4
p sjk2.6o C-2  # 5
p sjk2.cdi A-4

p B-9 a17.1i
s a17.op1 ε
s a17.cc1 C
p B-8 a18.1i
s a18.op1 ε
s a18.cc1 C
p C-1 a19.1i
s a19.op1 ε
s a19.cc1 C
p C-2 a20.1i
s a20.op1 ε
s a20.cc1 C

set a1 1
p a1.A 1
p A-1 a1.1i
s a1.op1 A
set a2 95 
p a2.A 1
p A-5 a2.1i
s a2.op1 A

p 1 ad.dp.1.2
p ad.dp.1.2 D-2
p 1 ad.dp.2.1
p ad.dp.2.1 D-1

p i.io A-1
p A-1 a3.5i
s a3.op5 0
s a3.cc5 0
s a3.rp5 1
p a3.5o A-2
p A-2 a3.6i
s a3.op6 0
s a3.cc6 0
s a3.rp6 1
p a3.6o A-3
p A-3 a3.7i
s a3.op7 0
s a3.cc7 0
s a3.rp7 1
p a3.7o A-4
p A-4 a3.8i
s a3.op8 0
s a3.cc8 0
s a3.rp8 1
p a3.8o A-5

p pa.1.sa.1 A-1
p pa.1.sb.1 A-10
p pa.1.sa.2 A-5
p pa.1.sb.2 A-10

b i
