# Fig 11-3: Isolation of programs through the use of a pulse amplifier
# This is pulseamp3.e but without a pulse amp to isolate 6-10 from 6-11.
# a1.5i will be triggered repeatedly and it'll go crazy and keep adding.
p 6-10 6-11
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
