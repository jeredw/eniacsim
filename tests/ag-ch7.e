# Quadratic Legrangian Interpolation
# Fig 7-3/Table 7-4, Adele Goldstine's Technical Manual

# Multiplier
p a9.S 8
p 3 ad.s.1.9
p ad.s.1.9 a9.α
p 3 ad.s.2.4
p ad.s.2.4 a9.β
p a10.S 7
p 2 a10.α
p m.lhppi 10
p m.rhppi 9
p a11.A 7
p 10 a11.α
p 8 a11.β
p a13.A 1
p 9 a13.α
p 7 a13.β
p m.Rα a9.1i
p m.Rβ a9.2i
p m.RS 3-1
p m.DS 3-3
p m.F 3-2
p 3-1 a9.3i
p m.Dα a10.1i
p 3-3 a10.2i
p 1-6 m.1i
p m.1o 1-8
p 1-10 m.2i
p m.2o 2-1
p 2-1 m.3i
p m.3o 2-4
p 2-4 m.4i
p m.4o 2-6
p 2-6 m.5i
p 3-2 a11.1i
p 3-1 a11.2i
p 3-2 a13.1i
p 3-3 a13.2i
p m.AC a13.3i
s a9.op1 α
s a9.op2 β
s a9.op3 S
s a10.op1 α
s a10.op2 S
s m.ieracc1 α
s m.icandacc1 α
s m.icandcl1 C
s m.sf1 6
s m.place1 2
s m.prod1 AC
s m.ieracc2 0
s m.iercl2 C
s m.icandacc2 α
s m.icandcl2 C
s m.sf2 6
s m.place2 2
s m.prod2 AC
s m.ieracc3 β
s m.iercl3 C
s m.icandacc3 α
s m.icandcl3 C
s m.sf3 6
s m.place3 6
s m.prod3 0
s m.ieracc4 β
s m.iercl4 C
s m.icandacc4 α
s m.icandcl4 C
s m.sf4 0
s m.place4 6
s m.prod4 0
s m.ieracc5 β
s m.iercl5 C
s m.icandacc5 α
s m.icandcl5 C
s m.sf5 0
s m.place5 6
s m.prod5 0
s a11.op1 A
s a11.cc1 C
s a11.op2 β
s a13.op1 β
s a13.op2 β
s a13.op3 A
s a13.cc3 C

# Accs 15-18
p a15.A 2
p 3 a15.α
p 1 ad.d.1.4
p ad.d.1.4 a15.β
p 1-7 a15.1i
p 1-2 a15.5i
p a15.5o 1-3
p 2-1 a15.6i
p a15.6o 2-2
p 2-4 a15.7i
p a15.7o 2-5
s a15.op1 α
s a15.op5 0
s a15.rp5 4
s a15.op6 β
s a15.rp6 6
s a15.op7 A
s a15.cc7 C
s a15.rp7 6

p a16.A 2
p a16.S 2
p 3 a16.α
p 3 ad.s.3.9
p ad.s.3.9 a16.β
p 2 a16.γ
p 1 ad.d.2.4
p ad.d.2.4 a16.δ
p 2-1 a16.1i
p 2-3 a16.2i
p 2-4 a16.3i
p 2-6 a16.4i
p 1-4 a16.5i
# p a16.5o 1-5
p a16.5o a16.12i
p a16.12o 1-5
p 1-8 a16.6i
p a16.6o 1-9
s a16.op1 A
s a16.op2 β
s a16.op3 γ
s a16.op4 S
s a16.cc4 C
s a16.op5 α
# s a16.rp5 4
s a16.rp5 1
s a16.op6 δ
s a16.rp6 3
s a16.op12 0
s a16.rp12 3

p a17.A 2
p a17.S 3
p 3 ad.s.4.4
p ad.s.4.4 a17.α
p 2 ad.s.5.4
p ad.s.5.4 a17.β
p 1-3 a17.1i
p 1-4 a17.2i
p 1-5 a17.3i
p 2-3 a17.4i
p 1-6 a17.5i
p a17.5o a17.6i
p 1-7 a17.7i
p 1-9 a17.8i
p 1-10 a17.9i
p a17.9o 1-11
p 2-2 a17.10i
# p a17.10o 2-3
p a17.10o a17.11i
p a17.11o 2-3
s a17.op1 α
s a17.op2 S
s a17.op3 α
s a17.op4 S
s a17.cc4 C
s a17.op5 A
s a17.cc5 C
s a17.rp5 4
s a17.op6 β
s a17.rp6 1
s a17.op7 S
s a17.rp7 1
s a17.op8 β
s a17.rp8 1
s a17.op9 A
s a17.cc9 C
s a17.rp9 2
s a17.op10 γ
s a17.cc10 C
s a17.rp10 1
s a17.op11 0
s a17.rp11 1

p a18.A 3
p 1 a18.α
p 2-10 a18.1i
p 2-11 a18.2i
p 1-6 a18.3i
p 1-1 a18.5i
p a18.5o 1-2
s a18.op1 A
s a18.op2 A
s a18.cc2 C
s a18.op3 A
s a18.op5 α
s a18.rp5 1

# Function Tables
p f2.A 3
p f2.B 2
p 3 ad.s.6.-1
p ad.s.6.-1 f2.arg
p f2.NC 2-10
p f2.C 2-11
p 1-2 f2.1i
p f2.1o 1-4
p 1-4 f2.2i
p f2.2o 1-6
p 1-6 f2.3i
p f2.3o 1-7
p 1-7 f2.4i
p f2.4o 1-10
s f2.op1 S0
s f2.cl1 NC
s f2.rp1 1
s f2.op2 A+1
s f2.cl2 NC
s f2.rp2 1
s f2.op3 S0
s f2.cl3 NC
s f2.rp3 1
s f2.op4 A+1
s f2.cl4 NC
s f2.rp4 1
s f2.mpm1 P
s f2.mpm2 P
s f2.A1d D
s f2.A2d D
s f2.A3d D
s f2.A4d D
s f2.B1d D
s f2.B2d D
s f2.B3d D
s f2.B4d D
s f2.A10s S
s f2.B10s S

p f3.A 3
p 3 ad.s.7.-3
p ad.s.7.-3 f3.arg
p f3.NC 2-10
p f3.C 2-11
p 1-11 f3.1i
p 2-2 f3.2i
p 2-5 f3.3i
s f3.op1 A+1
s f3.cl1 NC
s f3.rp1 1
s f3.op2 A0
s f3.cl2 NC
s f3.rp2 1
s f3.op3 A-1
s f3.cl3 C
s f3.rp3 1
s f3.mpm1 T
s f3.mpm2 T
s f3.A1d D
s f3.A2d D
s f3.A3d D
s f3.A4d D
s f3.B1d D
s f3.B2d D
s f3.B3d D
s f3.B4d D

l tests/legrange
l tests/cosine

# load the argument x into Acc 18
p c.o 1
p i.io 1-1
p 1-1 c.25i
s c.s25 Jlr
s c.j1 2
s c.j2 7
s c.j3 3
s c.j4 5
s c.j5 8
s c.j6 0
s c.j7 0
s c.j8 0
s c.j9 0
s c.j10 0

b i
