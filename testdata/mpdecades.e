# Test decade carry from decade direct inputs.
set a1 -9999999001

# Associate decades 17-14 with stepper C.
# Advance stepper C stage after 1000 counts.
s p.a14 C
s p.d17s1 1
s p.d16s1 0
s p.d15s1 0
s p.d14s1 0
# Set a nonzero limit for stage 2 so that we can tell stepper C has advanced.
s p.d14s2 9
s p.cC 2

# Wire together a1 and a2 to count up from -1000.
p a1.A 1
p a2.α 1
p a2.A 2
p a1.α 2

p 1-1 a1.5i
s a1.op5 A
s a1.cc5 C
s a1.rp5 1
p a1.5o 1-2

p 1-1 a2.1i
s a2.op1 α
s a2.cc1 0

p 1-2 a2.2i
s a2.op2 A
s a2.cc2 C

p 1-2 a1.6i
s a1.op6 α
s a1.cc6 C
s a1.rp6 1

# Retrigger 1-1 on send from a2 as long as sign is M.
p 2 ad.dp.1.11
p ad.dp.1.11 a20.12i
s a20.op12 0
s a20.cc12 0
s a20.rp12 1
p a20.12o 1-1

# Directly count at decade 14.
p 1-1 p.14di

p i.Io 1-1
b i