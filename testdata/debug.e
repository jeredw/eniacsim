p i.io 1-1

set a20 -2099920999
p debug.assert.1 1-1
s debug.assert.1 a20~M20xxx20xxx
p debug.dump.1 1-1
s debug.dump.1 a20

p debug.assert.2 1-2
s debug.assert.2 a20~x90xxxxxxxx

s a20.op11 0
p 1-1 a20.11i
p a20.11o 1-2
p 1-2 a20.12i
p a20.12o 1-3

p debug.bp.1 1-3

b i