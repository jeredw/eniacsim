# Test the order selector described in Clippinger '48 "A Logical Coding System
# Applied to the ENIAC"

p os.A 2
p os.B 3
p os.o 4
p os.i 2-1
p os.Ri 2-1
p os.Ro 1-10

p a6.A 1
p 1-2 a6.1i
s a6.op1 A

# f1 transmits row indexed by data trunk 1 on 1-1.
p 1 f1.arg
p 1-1 f1.1i
p f1.NC 1-2
p f1.A 2
p f1.B 3
s f1.op1 A0
s f1.cl1 NC
s f1.rp1 7
s f1.mpm1 P
s f1.RA42L6 1
s f1.RA42L5 0
s f1.RA42L4 2
s f1.RA42L3 0
s f1.RA42L2 3
s f1.RA42L1 0
s f1.RB42L6 4
s f1.RB42L5 0
s f1.RB42L4 5
s f1.RB42L3 0
s f1.RB42L2 6
s f1.RB42L1 0
# Dummy program triggers reading f1 A/B when ready.
p 1-1 a1.5i
p a1.5o 1-3
s a1.op5 0
s a1.rp5 4

# Read two digits in turn into a10-a16.
p a10.α 4
p a11.α 4
p a12.α 4
p a13.α 4
p a14.α 4
p a15.α 4
p a16.α 4
s a10.op5 α
s a11.op5 α
s a12.op5 α
s a13.op5 α
s a14.op5 α
s a15.op5 α
s a16.op5 α
p 1-3 a10.5i
p a10.5o 1-4
p 1-4 a11.5i
p a11.5o 1-5
p 1-5 a12.5i
p a12.5o 1-6
p 1-6 a13.5i
p a13.5o 1-7
p 1-7 a14.5i
p a14.5o 1-8
p 1-8 a15.5i
p a15.5o 1-9
p 1-9 a16.5i

# Buffer individual read enables onto the enable for the order selector.
p pa.1.sa.1 1-3
p pa.1.sb.1 2-1
p pa.1.sa.2 1-4
p pa.1.sb.2 2-1
p pa.1.sa.3 1-5
p pa.1.sb.3 2-1
p pa.1.sa.4 1-6
p pa.1.sb.4 2-1
p pa.1.sa.5 1-7
p pa.1.sb.5 2-1
p pa.1.sa.6 1-8
p pa.1.sb.6 2-1
p pa.1.sa.7 1-9
p pa.1.sb.7 2-1

# a20 will increment when os ring emits an overflow pulse.
p a20.ε 5
p 1-10 a20.1i
s a20.op1 ε
s a20.cc1 C

set a6 42

p i.io 1-1
b i
