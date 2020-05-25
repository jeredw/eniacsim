# Test ten position stepper as described in Clippinger '48 "A Logical Coding
# System Applied to the ENIAC"
#
# Decode a two digit number as illustrated in section 2, "Conversion of digit
# pulses to program pulses".  The expected decode output is wired to a program
# that increments a20, so a20 is nonzero if decoding works as expected.
# Incorrect decode outputs are wired to increment a19.

p st.i 1-1
p st.cdi 1-2
p st.di 1
p st.1o 4-4
p st.2o 4-4
p st.3o 4-4
p st.4o 4-4
p st.5o p.Di  # NB stage 5 is really 4
p st.6o 4-4
p st.7o 4-4
p st.8o 4-4
p st.9o 4-4
p st.10o 4-4

p p.Dcdi 1-1
p p.Ddi 2
p p.D1o 4-5
p p.D2o 4-5
p p.D3o 4-2
p p.D4o 4-5
p p.D5o 4-5
p p.D6o 4-5
s p.cD 6
s p.d13s1 9

# For simplicity emit tens digit from a1 and then ones digit from a2.
p a1.A 1
p 1-1 a1.5i
p a1.5o 1-2
s a1.op5 A
p a2.A 2
p 1-2 a2.1i
s a2.op1 A

# 4-5 means the mp decode was wrong.
p a18.α 3
p 4-5 a18.1i
s a18.op1 α
s a18.cc1 C

# 4-4 means the tenstepper decode was wrong.
p a19.α 3
p 4-4 a19.1i
s a19.op1 α
s a19.cc1 C

# 4-2 is the correct decode result.
p a20.α 3
p 4-2 a20.1i
s a20.op1 α
s a20.cc1 C

set a1 4
set a2 2

p i.io 1-1
b i
