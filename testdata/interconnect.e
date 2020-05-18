set a1 9999999999
set a2 9999999999

# Configure a1||a2 as a 20-digit accumulator with a1 holding the most
# significant digits.
p a1.il1 a1.il2
p a1.ir1 a2.il1
p a1.ir2 a2.il2

# Increment the accumulator pair.
p a2.5i 1-1
p a2.5o 1-2
s a2.op5 α
s a2.cc5 C

# Then increment again (controlled from a1).
p a1.5i 1-2
p a1.5o 1-3
s a1.op5 α
s a1.cc5 C

p i.io 1-1
b i
