# Fig 11-2: Bidirectional communication in pulse amplifier connected trays
# Connect trays 1 and 2 via pa.1 and again via pa.2.
p pa.1.sa 1
p pa.1.sb 2
p pa.2.sa 2
p pa.2.sb 1
p a1.α 1
p a1.β 2
p a2.A 1
p a2.S 2

# 1-1. a2 transmit S, a1 receive α (-42)
# 1-2. a2 transmit S, a1 receice β (-42)
# 1-3. a2 transmit A, a1 receive α (+42)
# 1-4. a2 transmit A, a1 receice β (+42)

p a1.5i 1-1
p a1.5o 1-2
p a1.6i 1-2
p a1.6o 1-3
p a1.7i 1-3
p a1.7o 1-4
p a1.8i 1-4
s a1.op5 α
s a1.op6 β
s a1.op7 α
s a1.op8 β

p a2.5i 1-1
p a2.6i 1-2
p a2.7i 1-3
p a2.8i 1-4
s a2.op5 S
s a2.op6 S
s a2.op7 A
s a2.op8 A

p i.io 1-1

set a1 0
set a2 42
b i
