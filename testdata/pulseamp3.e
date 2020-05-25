# Fig 11-3: Isolation of programs through the use of a pulse amplifier
# R1 = i.io, R2 = a1.5o
# P1 = a1.5i, P2 = a2.5i
# The pulse amp isolates 6-10 from 6-11, so P2 is activated twice and P1 is
# activated once (6-11 doesn't back-feed into 6-10).
p 6-10 pa.1.sa.1
p pa.1.sb.1 6-11

p a1.α 1
p 6-10 a1.5i
p a1.5o 6-11
s a1.op5 α

p a2.α 1
p 6-11 a2.5i
s a2.op5 α

p a20.A 1
p 6-11 a20.1i
s a20.op1 A

set a20 42

p i.io 6-10
b i
