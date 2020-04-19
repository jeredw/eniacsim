# Set-up for stimulating program Pi
#
# Master programmer, Acc 12 and Acc 14 from
# Figure 10-2

s p.a20 B
s p.cA 6
p 2-2 p.Adi
p 2-5 p.Ai
p 2-6 p.Acdi
p p.A1o 3-5
p p.A2o 3-6
p p.A3o 3-7
p p.A4o 3-8
p p.A5o 3-9

s p.a12 D
s p.cE 2
s p.d11s1 1
s p.d11s2 1
p 2-1 p.Edi
p 2-3 p.Ei
p 2-6 p.Ecdi
p p.E1o 2-4
p p.E2o 2-5

s p.a10 G
s p.cF 6
p 2-2 p.Fdi
p 2-4 p.Fi
p 2-6 p.Fcdi
p p.F1o 3-1
p p.F2o 3-2
p p.F3o 3-3
p p.F4o 3-4
p p.F6o 3-10

p a12.A 1
p 1-1 a12.1i
s a12.op1 A
p 2-4 a12.5i
p a12.5o 2-6
s a12.op5 0
s a12.rp5 2
p 2-5 a12.6i
p a12.6o 2-6
s a12.op6 0
s a12.rp6 2
s a12.sf 10

p 1-1 a14.5i
p a14.5o 1-2
s a14.op5 α
s a14.rp5 1
p 1-2 a14.6i
p a14.6o 2-3
s a14.op6 A
s a14.rp6 1
p 1 ad.d.1.-7
p ad.d.1.-7 a14.α
p a14.A 2
p 2 ad.dp.1.4
p 2 ad.dp.2.3
p ad.dp.1.4 2-1
p ad.dp.2.3 2-2
s a14.sf 7

p 3-1 a1.5i
s a1.op5 ε
s a1.rp5 1
s a1.cc5 C
p 3-2 a2.5i
s a2.op5 ε
s a2.rp5 1
s a2.cc5 C
p 3-3 a3.5i
s a3.op5 ε
s a3.rp5 1
s a3.cc5 C
p 3-4 a4.5i
s a4.op5 ε
s a4.rp5 1
s a4.cc5 C
p 3-5 a5.5i
s a5.op5 ε
s a5.rp5 1
s a5.cc5 C
p 3-6 a6.5i
s a6.op5 ε
s a6.rp5 1
s a6.cc5 C
p 3-7 a7.5i
s a7.op5 ε
s a7.rp5 1
s a7.cc5 C
p 3-8 a8.5i
s a8.op5 ε
s a8.rp5 1
s a8.cc5 C
p 3-9 a9.5i
s a9.op5 ε
s a9.rp5 1
s a9.cc5 C
p 3-10 a10.5i
s a10.op5 ε
s a10.rp5 1
s a10.cc5 C

# Initialization
p i.io 1-11
p c.o 1
p 1-11 c.25i
p c.25o 1-1
s c.s25 Jlr
s c.j1 3
s c.j2 2
s c.j3 3
p 1 a12.α
p 1-11 a12.2i
s a12.op2 α

b i
