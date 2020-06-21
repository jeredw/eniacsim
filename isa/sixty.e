# Generated by noweb, do not edit.  Use make to rebuild from sixty.nw.
# Input file for eniacsim (https://github.com/jeredw/eniacsim)
# Initiate pulse triggers A-1
p i.Io A-1

# Pulse amplifiers clear order selector and ft selector
# (A-1 -> A-3, A-1 -> E-9)
p A-3 os.Ci

# Dummy program to start fetch next cycle
p A-1 a2.12i
s a2.op12 0
s a2.cc12 0
s a2.rp12 1
p a2.12o C-5
# Fetch 0: Select function table.
# Select the FT to read.
# Decode E-8=0,1,2 into E-11, F-1, or F-2.
p E-1 sft.i
p E-8 sft.di
p sft.1o E-11
p sft.2o F-1
p sft.3o F-2

# Send PC on trunk 5.
p J-2 a6.5i
s a6.op5 A
s a6.cc5 0
s a6.rp5 1
p a6.5o E-9

# Await instruction data.
p C-5 a20.12i
s a20.op12 0
s a20.cc12 0
s a20.rp12 5
p a20.12o E-10
# Fetch 1: Initiate function table read.
# Clear FT selector for next fetch.
p E-9 sft.cdi

# Initiate instruction or operand read.
p E-11 f1.1i
s f1.op1 A-2
s f1.cl1 NC
s f1.rp1 1
p f1.NC J-4

p F-1 f2.1i
s f2.op1 A-2
s f2.cl1 NC
s f2.rp1 1
p f2.NC J-4

p F-2 f3.1i
s f3.op1 A-2
s f3.cl1 NC
s f3.rp1 1
p f3.NC J-4
# Fetch 2: Send function table argument.
# Send value for FT argument.
# Note this is also used by the 6(11,10,9) and 6(8,7) sequences.
p J-4 a6.6i
s a6.op6 A
s a6.cc6 0
s a6.rp6 1
p a6.6o F-3
# Fetch 3: Step order selector ring.
# Step the instruction ring counter.
p F-3 os.Ri
p os.Ro F-4  # overflow -> inc PC

# Dummy program to trigger Fetch 4.
p F-3 i.Ci4
p i.Co4 F-5
# Fetch 4: Clear decoder steppers and increment PC.
# Clear steppers prior to decode
p F-5 st.cdi
p F-5 p.Acdi
p F-5 p.Bcdi
p F-5 p.Ccdi
p F-5 p.Dcdi
p F-5 p.Ecdi
p F-5 p.Fcdi
p F-5 p.Gcdi
p F-5 p.Hcdi
p F-5 p.Jcdi
p F-5 p.Kcdi

# Increment PC if order selector wrapped
p F-4 a6.1i
s a6.op1 ε
s a6.cc1 C
# Dummy to enable order selector output next cycle.
p F-5 i.Ci1
p G-8 i.Co1
# Fetch 5: Decode instruction data.
# Order selector passes FT data.
p G-8 os.i

# Decode tens digit and trigger corresponding MP stepper input.
# Since only B-2 and B-5 are documented by available sources, just wire up
# the other ten stage stepper outputs directly.
p E-10 st.i
p F-6 st.di
p st.1o B-2
p B-2 p.Ai
p st.2o B-5
p B-5 p.Bi
p st.3o p.Ci
p st.4o p.Di
p st.5o p.Ei
p st.6o p.Fi
p st.7o p.Gi
p st.8o p.Hi
p st.9o p.Ji
p st.10o p.Ki

# Decode ones digit.
p E-2 p.Adi
p E-2 p.Bdi
p E-2 p.Cdi
p E-2 p.Ddi
p E-2 p.Edi
p E-2 p.Fdi
p E-2 p.Gdi
p E-2 p.Hdi
p E-2 p.Jdi
p E-2 p.Kdi

# Steppers use all 6 positions.
s p.cA 6
s p.cB 6
s p.cC 6
s p.cD 6
s p.cE 6
s p.cF 6
s p.cG 6
s p.cH 6
s p.cJ 6
s p.cK 6
# Set decade switches to nonzero values so that steppers don't cycle in between
# cdi and the decode cycle.
s p.d1s1 1
s p.d2s1 1
s p.d3s1 1
s p.d4s1 1
s p.d5s1 1
s p.d6s1 1
s p.d7s1 1
s p.d8s1 1
s p.d9s1 1
s p.d10s1 1
s p.d11s1 1
s p.d12s1 1
s p.d13s1 1
s p.d14s1 1
s p.d15s1 1
s p.d16s1 1
s p.d17s1 1
s p.d18s1 1
s p.d19s1 1
s p.d20s1 1
# Fetch 6: Wait for MP decoder.
# Waiting for MP decoder to propagate stepper input to output.
p p.A2o S-1   # 1l  (01)
p p.A3o S-2   # 2l  (02)
p p.A4o S-3   # 3l  (03)
p p.A5o S-4   # 4l  (04)
p p.A6o S-5   # 5l  (05)
p p.B1o S-6   # 7l  (10)
p p.B2o L-7   # 8l  (11)
p p.B3o L-8   # 9l  (12)
p p.B4o L-9   # 10l (13)
p p.B5o L-10  # 11l (14)
p p.B6o L-11  # 12l (15)
p p.C1o C-7   # 13l (20)  NB a13 += a15, doesn't clear first
p p.C2o H-1   # 14l (21)
p p.C3o H-2   # 16l (22)
p p.C4o H-3   # 17l (23)
p p.C5o H-4   # 18l (24)
p p.C6o H-5   # 19l (25)
p p.D2o V-9   # 1t  (31)
p p.D3o S-7   # 2t  (32)
p p.D4o S-8   # 3t  (33)
p p.D5o S-9   # 4t  (34)
p p.D6o S-10  # 5t  (35)
p p.E1o S-11  # 7t  (40)
p p.E2o C-2   # 8t  (41)
p p.E3o L-1   # 9t  (42)
p p.E4o L-2   # 10t (43)
p p.E5o L-3   # 11t (44)
p p.E6o L-4   # 12t (45)
p p.F1o L-5   # 13t (50)  NB clears after sending
p p.F2o L-6   # 14t (51)
p p.F3o C-9   # 16t (52)
p p.F4o H-6   # 17t (53)
p p.F5o H-7   # 18t (54)
p p.G5o H-8   # 19t (64)
p p.A1o C-1   # C   (00)  (also the general xl)
p p.D1o E-5   # X   (30)  "15 cycles"
p p.F6o E-3   # ÷   (55)  ~75 cycles
p p.G1o B-1   # √   (60)  ~75 cycles
p p.G4o V-3   # M   (63)
p p.G6o H-10  # DS  (65)
p p.J6o C-10  # Sh' (85)  20 cycles
p p.K1o C-11  # Sh  (90)  20 cycles
#p p.K2o C-8  # 20l (91a)
#p p.H4o E-4  # 6l  (73a) 9 cycles
#p p.H5o V-1  # 6t  (74a)
# Table 2.I has "H-0" for order 92a. Assume this is a typo for H-9 which is the
# only H-line missing, and makes sense from pulse amp grouping.
#p p.K3o H-9  # 20t  (92a)
#p p.K4o V-8  # N3D8 (93a) 20 cycles
#p p.K5o ???  # N3D6 (94a) 20 cycles
#p p.K6o ???  # N6D6 (95a) 26 cycles
p p.H6o G-2   # C.T.(75)  14/8 cycles
p p.H4o E-6   # 6R3 (73)  13 cycles
p p.H5o E-7   # 6R6 (74)  13 cycles
p p.G2o O-1   # Pr. (61)  60 cards / min
p p.G3o O-2   # Rd. (62)  100 cards / min
p p.H3o H-11  # F.T.(72)  13 cycles
p p.H1o V-5   # N2D (70)  14 cycles
p p.K5o V-6   # N4D (94)  20 cycles
p p.K6o V-7   # N6D (95)  26 cycles
p p.J1o D-7   # AB  (80)
p p.J2o D-8   # CD  (81)
p p.J3o D-9   # EF  (82)
p p.J4o D-10  # GH  (83)
p p.J5o D-11  # JK  (84)
#p p.H2o      # Halt       (71)  NB doesn't decode to anything, so stops
p p.K2o C-3   # 18<->20    (91)  9 cycles
p p.K3o V-2   # 6(11,10,9) (92)
p p.K4o V-4   # 6(8,7)     (93)
#  1    most accumulators transmit on 1
#  2    most accumulators receive on 2
#  3    ft data A
#  4    ft data B
#  5    ft argument
#  6,7  multiplier partial product digits (exclusive)
#  8=2  multiplier correction terms (shared)
p 8 2
#  9=1  multiplier correction terms (shared)
p 9 1
#  10   divider/square rooter answer (exclusive)
#  11=1 divider/square rooter shift (shared)
p 11 1

# Trunk 1 - accumulators, constants
p 1 c.o     # Constant data for C.T.

# Trunk 2 - accumulators, instructions
p 2 os.o    # Instruction to decode, and immediate operand for NxD

# Trunk 3 - ft data A
p 3 os.A    # Data for fetch and F.T.
p 3 f1.A
p 3 f2.A
p 3 f3.A

# Trunk 4 - ft data B
p 4 os.B    # Data for fetch and F.T.
p 4 f1.B
p 4 f2.B
p 4 f3.B

# Trunk 5
p 5 f1.arg  # Argument for F.T., PC
p 5 f2.arg
p 5 f3.arg

# Trunk 6/7 - multiplier partial products
p 6 m.lhppI
p 7 m.rhppI

# Trunk 10 - divider/square rooter answer
p 10 d.ans
# Accumulator 1
p 2 a1.α
p 1 a1.β
p ad.permute.4 a1.γ  # For 6(11,10,9)
p ad.permute.6 a1.δ  # For 6(8,7)
p 1 a1.A
p 1 a1.S  # Save to a13

# Accumulator 2
p 2 a2.α
p 1 a2.β
p ad.permute.20 a2.δ  # Sh  01 answer
p ad.permute.30 a2.ε  # Sh' 01 residual
p 1 a2.A

# Accumulator 3
p 2 a3.α
p 1 a3.β
p ad.permute.21 a3.δ  # Sh  02 answer
p ad.permute.31 a3.ε  # Sh' 02 residual
p 1 a3.A

# Accumulator 4 - quotient
p 10 a4.α  # Divider/square rooter answer
p 2 a4.β
p 1 a4.γ
p ad.permute.22 a4.δ  # Sh  03 answer
p ad.permute.32 a4.ε  # Sh' 03 residual
p 1 a4.A

# Accumulator 5 - numerator
p 2 a5.α
p 11 a5.γ  # Divider shift (1=11)
p ad.permute.23 a5.δ  # Sh  04 answer
p ad.permute.33 a5.ε  # Sh' 04 residual
p 11 a5.A  # Divider shift (1=11)

# Accumulator 6 - PC
p 2 a6.α
p 1 a6.β
p ad.permute.3 a6.γ  # For 6(11,10,9)
p ad.permute.5 a6.δ  # For 6(8,7)
#p 0 a6.ε  # NB ε is used as a dummy for increment
p 5 a6.A

# Accumulator 7 - denominator
p 2 a7.α
p 1 a7.β
p 10 a7.γ  # Divider answer
p ad.permute.24 a7.δ  # Sh  05 answer
p ad.permute.34 a7.ε  # Sh' 05 residual
p 11 a7.A  # Divider shift (1=11)
p 11 a7.S  # Divider shift (1=11)

# Accumulator 8 - F.T. argument
p 2 a8.α
p 5 a8.A

# Accumulator 9 - shift
p ad.s.1.1 a9.α
p 2 a9.β
p 1 a9.γ
p ad.permute.29 a9.δ  # Sh  95 answer
p ad.permute.39 a9.ε  # Sh' 95 residual
p 11 a9.A  # Divider shift (1=11)

# Accumulator 10
p 2 a10.α
p ad.permute.7 a10.β  # For 6R3
p ad.permute.8 a10.γ  # For 6R3
p ad.permute.11 a10.δ  # For C.T.
p 1 a10.A
p 2 a10.S  # Save to a13

# Accumulator 11 - ier
p 2 a11.α
p ad.permute.1 a11.β  # F.T. data A
p ad.permute.12 a11.γ  # For DS
p 1 a11.δ
p 1 a11.A
p 8 a11.S  # Multiplier correction (8=2)

# Accumulator 12 - icand
p 2 a12.α
p 1 a12.β
p 1 a12.A
p 9 a12.S  # Multiplier correction (9=1)

# Accumulator 13 - LHPP
p 6 a13.α  # Multiplier partial product
p 8 a13.β  # Multiplier correction (8=2)
p 1 a13.γ
p 1 a13.A
p 2 a13.S  # Restore temporaries

# Accumulator 14
p 2 a14.α
p 1 a14.β
p ad.permute.28 a14.δ  # Sh  94 answer
p ad.permute.38 a14.ε  # Sh' 94 residual
p 1 a14.A

# Accumulator 15
p 7 a15.α  # Multiplier partial product
p 9 a15.β  # Multiplier correction (9=1)
p ad.permute.2 a15.γ  # F.T. data B
p 5 ad.δ  # For 8t
p 2 a5.A

# Accumulator 16
p 2 a16.α
p 1 a16.β
p ad.permute.27 a16.δ  # Sh  93 answer
p ad.permute.37 a16.ε  # Sh' 93 residual
p 1 a16.A

# Accumulator 17
p 2 a17.α
p 1 a17.β 
p ad.permute.26 a17.δ  # Sh  92 answer
p ad.permute.36 a17.ε  # Sh' 92 residual
p 1 a17.A

# Accumulator 18
p 2 a18.α
p 1 a18.β
p ad.permute.13 a18.γ  # For N4D
p ad.permute.14 a18.δ  # For N6D
p 1 a18.A
p 2 a18.S  # Save to a13

# Accumulator 19
p 2 a19.α
p 1 a19.β
p ad.permute.25 a19.δ  # Sh  91 answer
p ad.permute.35 a19.ε  # Sh' 91 residual
p 1 a19.A

# Accumulator 20
p 2 a20.α
p 1 a20.β
p ad.permute.9 a20.γ  # For 6R6
p ad.permute.10 a20.δ  # For 6R6
p 1 a20.A
p 2 a20.S  # Save to a13
# Function table output A, shifted to the left for F.T.
p 3 ad.permute.1
s ad.permute.1 11,6,5,4,3,2,1,0,0,0,0
# Function table output B, shifted to the left for F.T.
p 4 ad.permute.2
s ad.permute.2 11,6,5,4,3,2,1,0,0,0,0
# For 6(11,10,9).
p 2 ad.permute.3
s ad.permute.3 11,2,1,0,0,0,0,0,0,0,0
p 5 ad.permute.4
s ad.permute.4 11,0,0,0,0,0,0,0,0,10,9
# For 6(8,7).
p 2 ad.permute.5
s ad.permute.5 0,0,0,2,1,0,0,0,0,0,0
p 5 ad.permute.6
s ad.permute.6 0,0,0,0,0,0,0,0,0,8,7
# For 6R3.
p 5 ad.permute.7
s ad.permute.7 11,10,9,8,7,6,5,4,0,0,0
p 2 ad.permute.8
s ad.permute.8 0,0,0,0,0,0,0,0,3,2,1
# For 6R6.
p 5 ad.permute.9
s ad.permute.9 11,10,9,8,7,0,0,0,0,0,0
p 2 ad.permute.10
s ad.permute.10 0,0,0,0,0,6,5,4,3,2,1
# For C.T., shift a6 digits 6-4 to position 3-1.
p 5 ad.permute.11
s ad.permute.11 11,10,9,8,7,0,0,0,6,5,4
# Delete sign for DS
p 1 ad.permute.12
s ad.permute.12 0,10,9,8,7,6,5,4,3,2,1
# Shift for N4D.
p 2 ad.permute.13
s ad.permute.13 0,0,0,0,0,0,0,2,1,0,0
# Shift for N6D.
p 2 ad.permute.14
s ad.permute.14 0,0,0,0,0,2,1,0,0,0,0
# For divider shifting
p 11 ad.s.1.1
# Left shift adapters for Sh on a15.
p 2 ad.permute.20
s ad.permute.20 11,9,8,7,6,5,4,3,2,1,0       # << 1
p 2 ad.permute.21                            
s ad.permute.21 11,8,7,6,5,4,3,2,1,0,0       # << 2
p 2 ad.permute.22                            
s ad.permute.22 11,7,6,5,4,3,2,1,0,0,0       # << 3
p 2 ad.permute.23                            
s ad.permute.23 11,6,5,4,3,2,1,0,0,0,0       # << 4
p 2 ad.permute.24                            
s ad.permute.24 11,5,4,3,2,1,0,0,0,0,0       # << 5
# Right arithmetic shift adapters for Sh on a15.
p 2 ad.permute.25
s ad.permute.25 11,11,10,9,8,7,6,5,4,3,2     # >> 1
p 2 ad.permute.26
s ad.permute.26 11,11,11,10,9,8,7,6,5,4,3    # >> 2
p 2 ad.permute.27
s ad.permute.27 11,11,11,11,10,9,8,7,6,5,4   # >> 3
p 2 ad.permute.28
s ad.permute.28 11,11,11,11,11,10,9,8,7,6,5  # >> 4
p 2 ad.permute.29
s ad.permute.29 11,11,11,11,11,11,10,9,8,7,6 # >> 5
# Sh' opposite of left shifts.
p 1 ad.permute.30
s ad.permute.30 11,11,11,11,11,11,11,11,11,11,10  # >> 9
p 1 ad.permute.31
s ad.permute.31 11,11,11,11,11,11,11,11,11,10,9   # >> 8
p 1 ad.permute.32
s ad.permute.32 11,11,11,11,11,11,11,11,10,9,8    # >> 7
p 1 ad.permute.33
s ad.permute.33 11,11,11,11,11,11,11,10,9,8,7     # >> 6
p 1 ad.permute.34
s ad.permute.34 11,11,11,11,11,11,10,9,8,7,6      # >> 5
# Sh' opposite of right shifts.
p 1 ad.permute.35
s ad.permute.35 11,1,0,0,0,0,0,0,0,0,0            # << 9
p 1 ad.permute.36
s ad.permute.36 11,2,1,0,0,0,0,0,0,0,0            # << 8
p 1 ad.permute.37
s ad.permute.37 11,3,2,1,0,0,0,0,0,0,0            # << 7
p 1 ad.permute.38
s ad.permute.38 11,4,3,2,1,0,0,0,0,0,0            # << 6
p 1 ad.permute.39
s ad.permute.39 11,5,4,3,2,1,0,0,0,0,0            # << 5
# Special digit adapters
# E-8 is the third digit of a6 or a8 from trunk 5.
p 5 ad.dp.1.3
p ad.dp.1.3 E-8
# E-2 selects the ones digit of the order from trunk 2.
p 2 ad.dp.2.1
p ad.dp.2.1 E-2
# F-6 selects the tens digit of the order from trunk 2.
p 2 ad.dp.3.2
p ad.dp.3.2 F-6
# Box 1: Listen orders (xl) -> C-1
# Note that C-1 is also the encoding for 00 (C) which clears a15.
# 6l needs special handling.
p pa.1.sa.1  S-1  # 1l
p pa.1.sb.1  C-1
p pa.1.sa.2  S-2  # 2l
p pa.1.sb.2  C-1
p pa.1.sa.3  S-3  # 3l
p pa.1.sb.3  C-1
p pa.1.sa.4  S-4  # 4l
p pa.1.sb.4  C-1
p pa.1.sa.5  S-5  # 5l
p pa.1.sb.5  C-1
p pa.1.sa.6  S-6  # 7l
p pa.1.sb.6  C-1
p pa.1.sa.7  L-7  # 8l
p pa.1.sb.7  C-1
p pa.1.sa.8  L-8  # 9l
p pa.1.sb.8  C-1
p pa.1.sa.9  L-9  # 10l
p pa.1.sb.9  C-1
p pa.1.sa.10 L-10 # 11l
p pa.1.sb.10 C-1
p pa.1.sa.11 L-11 # 12l
p pa.1.sb.11 C-1
p pa.2.sa.1  C-7  # 13l
p pa.2.sb.1  C-1
p pa.2.sa.2  H-1  # 14l
p pa.2.sb.2  C-1
p pa.2.sa.3  H-2  # 16l
p pa.2.sb.3  C-1
p pa.2.sa.4  H-3  # 17l
p pa.2.sb.4  C-1
p pa.2.sa.5  H-4  # 18l
p pa.2.sb.5  C-1
p pa.2.sa.6  H-5  # 19l
p pa.2.sb.6  C-1
#p pa.x.sa.y  C-8  # 20l
#p pa.x.sb.y  C-1
# Box 2: Talk orders (xt) -> D-3
# 6t and 8t need special handling because a6 and a8 transmit on a separate
# trunk to address the function table.
p pa.2.sa.7  V-9  # 1t
p pa.2.sb.7  D-3
p pa.2.sa.8  S-7  # 2t
p pa.2.sb.8  D-3
p pa.2.sa.9  S-8  # 3t
p pa.2.sb.9  D-3
p pa.2.sa.10 S-9  # 4t
p pa.2.sb.10 D-3
p pa.2.sa.11 S-10 # 5t
p pa.2.sb.11 D-3
p pa.3.sa.1  S-11 # 7t
p pa.3.sb.1  D-3
p pa.3.sa.2  L-1  # 9t
p pa.3.sb.2  D-3
p pa.3.sa.3  L-2  # 10t
p pa.3.sb.3  D-3
p pa.3.sa.4  L-3  # 11t
p pa.3.sb.4  D-3
p pa.3.sa.5  L-4  # 12t
p pa.3.sb.5  D-3
p pa.3.sa.6  L-5  # 13t
p pa.3.sb.6  D-3
p pa.3.sa.7  L-6  # 14t
p pa.3.sb.7  D-3
p pa.3.sa.8  C-9  # 16t
p pa.3.sb.8  D-3
p pa.3.sa.9  H-6  # 17t
p pa.3.sb.9  D-3
p pa.3.sa.10 H-7  # 18t
p pa.3.sb.10 D-3
p pa.3.sa.11 H-8  # 19t
p pa.3.sb.11 D-3
#p pa.x.sa.y  H-9  # 20t
#p pa.x.sb.y  D-3
# Box 3: Constant transfers (uv) -> D-4
p pa.4.sa.1  D-7  # AB
p pa.4.sb.1  D-4
p pa.4.sa.2  D-8  # CD
p pa.4.sb.2  D-4
p pa.4.sa.3  D-9  # EF
p pa.4.sb.3  D-4
p pa.4.sa.4  D-10 # GH
p pa.4.sb.4  D-4
p pa.4.sa.5  D-11 # JK
p pa.4.sb.5  D-4
# Box 4: Misc remaining 7 cycle ops (a1tmp) -> D-5
# These instructions use a1 as a temporary.
p pa.4.sa.6 V-2  # 6(11,10,9)
p pa.4.sb.6 D-5
p pa.4.sa.7 V-3  # M
p pa.4.sb.7 D-5
p pa.4.sa.8 V-4  # 6(8,7)
p pa.4.sb.8 D-5
# Talk or constant transfers -> B-3
# B-3 triggers a15 to receive on trunk 1.
p pa.4.sa.9 D-3  # xt
p pa.4.sb.9 B-3
p pa.4.sa.10 D-4  # uv
p pa.4.sb.10 B-3
# Dual of constant transfers -> J-3
p pa.4.sa.11 D-4  # uv
p pa.4.sb.11 J-3
# Dual of a1tmp -> B-4
p pa.5.sa.1 D-5  # a1tmp
p pa.5.sb.1 B-4
# C-5 triggers fetching _and decoding_ another instruction.
# C-1 (xl), D-3 (xt), D-4 (uv), D-5 (a1tmp) as well as C-2 (8t), V-1 (6t in
# alternate code), and H-10 (DS) all do so immediately upon decode.
p pa.5.sa.2 C-1  # xl
p pa.5.sb.2 C-5
p pa.5.sa.3 D-3  # xt
p pa.5.sb.3 C-5
p pa.5.sa.4 D-4  # uv
p pa.5.sb.4 C-5
p pa.5.sa.5 D-5  # a1tmp
p pa.5.sb.5 C-5
p pa.5.sa.6 C-2  # 8t
p pa.5.sb.6 C-5
#p pa.x.sa.y V-1  # 6t
#p pa.x.sb.y C-5
p pa.5.sa.7 H-10 # DS
p pa.5.sb.7 C-5
# D-6 triggers the fetch sequence separately from decode.
p pa.5.sa.8 D-1  # N2D triggered from N4D/N6D
p pa.5.sb.8 D-6
p pa.5.sa.9 C-6  # N2D/N4D/N6D/N3D8 use fetch
p pa.5.sb.9 D-6
p pa.5.sa.10 D-2  # N4D triggered from N6D
p pa.5.sb.10 D-6
p pa.5.sa.11 C-5  # fetch+decode -> fetch
p pa.5.sb.11 D-6
p pa.6.sa.1 C-11 # Sh/Sh' take a shift amount
p pa.6.sb.1 D-6
# J-2 and E-1 are duals of D-6
p pa.6.sa.2 D-6
p pa.6.sb.2 J-2  # PC update sequence
p pa.6.sa.3 D-6
p pa.6.sb.3 E-1  # FT read sequence
# C-6 is {N2D,N4D,N6D,N3D8} - instructions that use a18 temporary.
p pa.6.sa.4 V-5  # N2D
p pa.6.sb.4 C-6
p pa.6.sa.5 V-7  # N6D
p pa.6.sb.5 C-6
#p pa.x.sa.y V-8   # N3D8
#p pa.x.sb.y C-6
p pa.6.sa.6 V-6  # N4D
p pa.6.sb.6 C-6

# D-1 is dual of N2D, also triggered by N4D sequence.
p pa.6.sa.7 V-5  # N2D
p pa.6.sb.7 D-1
# D-2 is dual of N4D, also triggered by N6D sequence.
p pa.6.sa.8 V-6  # N4D
p pa.6.sb.8 D-2
# This connection is given in table 2.II without context.
# It's useful for the a1tmp sequence, so use it for that.
p pa.6.sa.9 C-4  # cycle 8 of a1tmp instructions
p pa.6.sb.9 J-1  # a15 send and clear

# Note C-11 is also the encoding for Sh.
p pa.6.sa.10 C-10 # Sh'
p pa.6.sb.10 C-11 # Sh

#p pa.x.sa.y C-8  # 20l
#p pa.x.sb.y K-11
# Reset sequence
p pa.6.sa.11 A-1   # initiate pulse
p pa.6.sb.11 A-3   # clear order selector
p pa.7.sa.1  A-1   # initiate pulse
p pa.7.sb.1  E-9   # clear ft selector
# a1tmp: Trigger a13 receive
p pa.7.sa.2 D-5   # a1tmp
p pa.7.sb.2 B-10   # a13 receive
# a1tmp: Trigger a13 send
p pa.7.sa.3 T-8    # a1tmp+11
p pa.7.sb.3 B-11    # a13 send and clear

# X: Trigger a15 to receive correction term
p pa.7.sa.5 M-3   # X+19
p pa.7.sb.5 B-3   # a15 receive correction term
# X: Trigger a15 to receive partial product
p pa.7.sa.6 M-4   # X+20
p pa.7.sa.6 B-11  # a13 send and clear
p pa.7.sa.7 M-4   # X+20
p pa.7.sb.7 B-3   # a15 receive lhpp

# ÷: Save lots of temporaries
p pa.7.sa.8 E-3   # ÷
p pa.7.sb.8 B-10   # a13 receive a4

p pa.7.sa.9 R-1   #
p pa.7.sb.9 J-1   # a15 send and clear
p pa.7.sa.10 R-2   #
p pa.7.sb.10 B-3   # a15 receive a9
p pa.7.sa.11 R-4   #
p pa.7.sb.11 J-1   # a15 send and clear

# xl instruction
# xl cycle 7: Clear aL
# Dummy program to delay C-1->J-1, i.e. xl->a15 send.
# Figure 3.3 has this on the constant unit.  Note that this puts junk on
# trunk 1 during xl cycle 7 and conflicts with a1:AC1.
p C-1 c.3i
p c.3o J-1

# Clear listener
# 1l
p S-1 a1.5i
s a1.op5 A  # NB this is AC1 in Figure 3.3 not 0C1.
s a1.cc5 C
s a1.rp5 1
p a1.5o O-8  # O-8 is given in Figure 3.3.
# To conserve program lines, wire aN.5o directly to aN.1i for N>1.  In practice
# this would have been done with U cables (ETM XI-8).
# 2l
p S-2 a2.5i
s a2.op5 A
s a2.cc5 C
s a2.rp5 1
p a2.5o a2.1i
# 3l
p S-3 a3.5i
s a3.op5 A
s a3.cc5 C
s a3.rp5 1
p a3.5o a3.1i
# 4l
p S-4 a4.5i
s a4.op5 A
s a4.cc5 C
s a4.rp5 1
p a4.5o a4.1i
# 5l
p S-5 a5.5i
s a5.op5 A
s a5.cc5 C
s a5.rp5 1
p a5.5o a5.1i
# 6l = E-4 is not implemented
# 7l 
p S-6 a7.5i
s a7.op5 A
s a7.cc5 C
s a7.rp5 1
p a7.5o a7.1i
# 8l
p L-7 a8.5i
s a8.op5 A
s a8.cc5 C
s a8.rp5 1
p a8.5o a8.1i
# 9l
p L-8 a9.5i
s a9.op5 A
s a9.cc5 C
s a9.rp5 1
p a9.5o a9.1i
# 10l
p L-9 a10.5i
s a10.op5 A
s a10.cc5 C
s a10.rp5 1
p a10.5o a10.1i
# 11l
p L-10 a11.5i
s a11.op5 A
s a11.cc5 C
s a11.rp5 1
p a11.5o a11.1i
# 12l
p L-11 a12.5i
s a12.op5 A
s a12.cc5 C
s a12.rp5 1
p a12.5o a12.1i
# 13l: Note that 13l does a13 += a15 rather than a13 = a15
p C-7 a13.5i
s a13.op5 0
s a13.cc5 0
s a13.rp5 1
p a13.5o a13.1i
# 14l
p H-1 a14.5i
s a14.op5 A
s a14.cc5 C
s a14.rp5 1
p a14.5o a14.1i
# 16l
p H-2 a16.5i
s a16.op5 A
s a16.cc5 C
s a16.rp5 1
p a16.5o a16.1i
# 17l
p H-3 a17.5i
s a17.op5 A
s a17.cc5 C
s a17.rp5 1
p a17.5o a17.1i
# 18l 
p H-4 a18.5i
s a18.op5 A
s a18.cc5 C
s a18.rp5 1
p a18.5o a18.1i
# 19l
p H-5 a19.5i
s a19.op5 A
s a19.cc5 C
s a19.rp5 1
p a19.5o a19.1i
# 20l = C-8 is not implemented
# xl cycle 8: Receive from a15
# a15 send and clear
p J-1 a15.1i
s a15.op1 A
s a15.cc1 C

# Receive from a15
p O-8 a1.1i
s a1.op1 α
# aN.1i for N>1 is wired up directly in xl cycle 7
s a2.op1 α
s a3.op1 α
s a4.op1 β
s a5.op1 α
# 6l is not implemented
s a7.op1 α
s a8.op1 α
s a9.op1 β
s a10.op1 α
s a11.op1 α
s a12.op1 α
s a13.op1 β
s a14.op1 α
s a16.op1 α
s a17.op1 α
s a18.op1 α
s a19.op1 α
# 20l is not implemented
# xt instruction
# Cycle 7: a15 receive from aX
p B-3 a15.2i
s a15.op2 β
# Special case for 8t: a15 receive on trunk 5
p C-2 a15.3i
s a15.op3 δ

# 1t
p V-9 a1.2i
s a1.op2 A
p S-7 a2.2i
s a2.op2 A
# 3t
p S-8 a3.2i
s a3.op2 A
# 4t
p S-9 a4.2i
s a4.op2 A
# 5t
p S-10 a5.2i
s a5.op2 A
# 6t = V-1 is not implemented
# 7t
p S-11 a7.2i
s a7.op2 A
# 8t
p C-2 a8.2i
s a8.op2 A
# 9t
p L-1 a9.2i
s a9.op2 A
# 10t
p L-2 a10.2i
s a10.op2 A
# 11t
p L-3 a11.2i
s a11.op2 A
# 12t
p L-4 a12.2i
s a12.op2 A
# 13t: Uniquely, 13t clears after sending
p L-5 a13.2i
s a13.op2 A
s a13.cc2 C
# 14t
p L-6 a14.2i
s a14.op2 A
# 16t
p C-9 a16.2i
s a16.op2 A
# 17t
p H-6 a17.2i
s a17.op2 A
# 18t
p H-7 a18.2i
s a18.op2 A
# 19t
p H-8 a19.2i
s a19.op2 A
# AB/CD/EF/GH/JK instructions
# Constant cycle 7: right digits to a15
# (Pulse amps trigger B-3 to add the constant to a15.)
p D-7 c.1i    # AB
s c.s1 Blr
p c.1o c.2i
p D-8 c.7i    # CD
s c.s7 Dlr
p c.7o c.8i
p D-9 c.13i   # EF
s c.s13 Flr
p c.13o c.14i
p D-10 c.19i  # GH
s c.s19 Hlr
p c.19o c.20i
p D-11 c.25i  # JK
s c.s25 Klr
p c.25o c.26i
# Clear a11 to receive a constant next cycle.
p J-3 a11.6i
s a11.op6 0
s a11.cc6 C
s a11.rp6 1
p a11.6o T-6
# Constant cycle 8: left digits to a11
# Receive constant digits.
p T-6 a11.3i
s a11.op3 δ
# c.2i, c.8i, ... are connected directly from c.1o, c.7o, etc.
s c.s2 Alr
s c.s8 Clr
s c.s14 Elr
s c.s20 Glr
s c.s26 Jlr
# Several instructions that use a1 as a temporary share the same sequence.
# Cycle 7: Save a1 in a13
# Transmit a1
p D-5 a1.6i
s a1.op6 A
s a1.cc6 C
s a1.rp6 1
p a1.6o C-4  # Trigger a15 to send in cycle 8
# D-5 -> B-10 triggers a13 receive
# Receive a13
p B-10 a13.3i
s a13.op3 γ
s a13.cc3 0
p D-5 a19.12i # Dummy to receive answer/restore a13
s a19.op12 0
s a19.cc12 0
s a19.rp12 3
p a19.12o T-7

# Cycle 8: C-4 sends a15; per-instruction work

# Cycle 9: per-instruction work

# Cycle 10: Receive answer
p T-7 a15.5i
s a15.op5 β
s a15.cc5 0
s a15.rp5 1
p a15.5o T-8

# Cycle 11: Restore a1 from a13
# T-8 -> B-11 is triggered by pulse amplifiers.
# Transmit a13
p B-11 a13.6i
s a13.op6 A
s a13.cc6 C
s a13.rp6 1
# Receive a1
p T-8 a1.3i
s a1.op3 β
s a1.cc3 0

# These are the steps that vary per instruction.
# 6(11,10,9)
# Cycle 7: Dummy to delay V-2
p V-2 a2.6i
s a2.op6 0
s a2.cc6 0
s a2.rp6 1
p a2.6o T-1

# Cycle 8: Add a15(11,2,1) to a6(11,10,9)
p T-1 a6.7i
s a6.op7 γ
s a6.cc7 0
s a6.rp7 1
p a6.7o T-2

# Cycle 9: Receive 6(11,10,9) in 1(11,2,1)
# Fetch J-4 will send a6 this cycle.
p T-2 a1.7i
s a1.op7 γ
s a1.cc7 0
s a1.rp7 1
p a1.7o T-3

# Cycle 10: Transmit a1
# (a15 receive is part of common a1tmp sequence)
p T-3 a1.8i
s a1.op8 A
s a1.cc8 C
s a1.rp8 1
# Cycle 7: Dummy to delay V-4
p V-4 a3.6i
s a3.op6 0
s a3.cc6 0
s a3.rp6 1
p a3.6o T-9

# Cycle 8: Add a15(2,1) to a6(8,7)
p T-9 a6.8i
s a6.op8 δ
s a6.cc8 0
s a6.rp8 1
p a6.8o T-10

# Cycle 9: Receive 6(8,7) in 1(2,1)
# Fetch J-4 will send a6 this cycle.
p T-10 a1.9i
s a1.op9 δ
s a1.cc9 0
s a1.rp9 1
p a1.9o T-11

# Cycle 10: Transmit a1
# (a15 receive is part of common a1tmp sequence)
p T-11 a1.10i
s a1.op10 A
s a1.cc10 C
s a1.rp10 1
# Cycle 7: Dummy to delay V-3
p V-3 a16.6i
s a16.op6 0
s a16.cc6 0
s a16.rp6 1
p a16.6o T-4

# Cycle 8: a1 gets a15
p T-4 a1.11i
s a1.op11 α
s a1.cc11 0
s a1.rp11 1
p T-4 a16.7i  # Dummy to wait for cycle 10.
s a16.op7 0
s a16.cc7 0
s a16.rp7 2
p a16.7o T-5

# Cycle 9: nop

# Cycle 10: Transmit a1 subtractively
# (a15 receive is part of common a1tmp sequence)
p T-5 a1.4i
s a1.op4 S
s a1.cc4 C
# Multiply
# The static multiplier wiring is set up for ier=a11 and icand=a12
p m.ier a11
p m.icand a12
p m.L a13
p m.R a15
s m.ieracc1 0
s m.iercl1 0
s m.icandacc1 α  # Trigger a12:α01 in cycle 8
s m.icandcl1 0
s m.sf1 off
s m.place1 10
s m.prod1 0

# Cycle 7: Start multiplication and clear icand
# Start multiplication
p E-5 m.1i
# Clear icand
p E-5 a12.6i
s a12.op6 0
s a12.cc6 C
s a12.rp6 1
p a12.6o M-1
# Trigger J-1 to transmit a15 to icand in cycle 8
p E-5 c.4i
p c.4o J-1  # NB uses trunk 1

# Cycle 8: Dummy program triggers a15:AC1

# Cycle 9..18: (multiply)

# Cycle 19: Add correction terms for signed arguments
# Correction from ier
p m.RS M-2
p M-2 a11.4i
s a11.op4 S
p M-2 a13.4i
s a13.op4 β
# Correction from icand
p m.DS M-3
p M-3 a12.3i
s a12.op3 S
# Pulse amplifiers trigger a15:β01 (M-3 -> B-3)

# Cycle 20: Combine partial products into a15
p m.F M-4
# Pulse amplifiers trigger a13:AC1 (M-4 -> B-11)
# Pulse amplifiers trigger a15:β01 (M-4 -> B-3)
p m.1o C-5  # Retrigger fetch sequence
# Divider/square rooter accumulator wiring
p d.quotient a4
p d.numerator a5
p d.denominator a7
p d.shift a9

# A few sequences are common to divide and square root.
# R-1: Transfer a15 to a5, then save a9 in a15.
# Cycle 0: Transfer a15 to a5.
# Pulse amplifiers trigger a15:AC1 (R-1 -> J-1)
p R-1 a5.7i
s a5.op7 α
s a5.cc7 0
s a5.rp7 1
p a5.7o R-2
# Cycle 1: Save a9 in a15.
p R-2 a9.6i
s a9.op6 A
s a9.cc6 C
s a9.rp6 1
# Pulse amplifiers trigger a15:β01 (R-2 -> B-3)
# R-3: Clear a9, then restore from a15.
# Cycle 0: Clear a9
p R-3 a9.7i
s a9.op7 0
s a9.cc7 C
s a9.rp7 1
p a9.7o R-4
# Cycle 1: Save a9 in a15.
# Pulse amplifiers trigger a15:AC1 (R-4 -> J-1)
p R-4 a9.8i
s a9.op6 γ
s a9.cc6 0
s a9.rp6 1
# Use divider input 1, transmitting arguments via external programs not
# divider-generated pulses due to the need to shuffle accumulators.
s d.nr1 0
s d.nc1 0
s d.dr1 0
s d.dc1 0
s d.pl1 D10
s d.ro1 RO
s d.an1 1
s d.il1 NI

# Cycle 7: Clear a5 and save a4 in a13
p E-3 a5.6i
s a5.op6 0
s a5.cc6 C
s a5.rp6 1
p a5.6o R-1  # Trigger R-1 sequence for cycle 8+9
p E-3 a4.3i
s a4.op3 A
s a4.cc3 C
# Pulse amplifiers trigger a13:γ01 (E-3 -> B-10)
# Dummy to continue division
p E-3 i.Ci2
p i.Co2 V-10

# Cycle 8+9: Sequence R-1 (Q-1 -> R-1)

# Cycle 10-89*: Divide (* e.g.)
p Q-4 d.1i
p d.1o Q-5

# Cycle 90: Clear a9 before restoring
p Q-5 a9.7i
p a9.7o Q-6

# Cycle 91: Restore a9
p Q-6 a9.8i
p a9.8o Q-7

# Cycle 92: Transfer quotient to a15
p Q-7 a4.6i
p a4.6o Q-8

# Cycle 93: Restore a4
p Q-8 a4.6i
s d.nr2 0
s d.nc2 0
s d.dr2 0
s d.dc2 0
s d.pl2 R10
s d.ro2 RO
s d.an2 4
s d.il2 NI
s d.da A
s d.ra A

p B-1 ...
# DS
# Cycle 7: Save a11
p H-10 a11.9i
s a11.op9 A
s a11.cc9 C
s a11.rp9 1
p a11.9o x1
# Pulse amplifiers trigger a13:γ01 (H-10 -> B-10)
p H-10 

# Cycle 8: Transmit a15, dropping sign [trunk 2]
# Pulse amplifiers trigger a15:AC1 (x1 -> J-1)
p x1 a11.10i
s a11.op10 γ
s a11.cc10 0
s a11.rp10 1
p a11.10o x2
# Trigger B-3 a15:β01 next cycle.
p x1 c.5i
p c.5o B-3  # Uses trunk 1 this cycle

# Cycle 9: Get a15 without sign.
p x2 a11.11i
s a11.op11 A
s a11.cc11 C
s a11.rp11 1
p a11.11o x3
# Dummy from cycle 8 triggers a15:β01

# Cycle 10: Restore a11
p x3 a11.12i
s a11.op12 β
s a11.cc12 0
s a11.rp12 1
# Cycle 7: (for all NxD instructions)
# Pulse amplifiers assert D-6 to fetch the operand.
# Save -a18 if triggered directly from decode.
p C-6 a18.2i
s a18.op2 S
s a18.cc2 C
#p C-6 a13.2i
#s a13.op2 β

# N6D cycle 7
# Await FT data
p V-7 a18.5i  # N6D
s a18.op5 0
s a18.cc5 0
s a18.rp5 5
p a18.5o .x1

# N6D cycle 12:
# Receive << 4 shifted digits from function table
p .x1 a18.6i
s a18.op6 δ
s a18.cc6 0
s a18.rp6 1
p a18.6o .x2
# Trigger N4D sequence via a dummy program
p .x1 a18.7i
s a18.op7 0
s a18.cc7 0
s a18.rp7 1
p a18.7o D-2

# N6D cycle 13:
# Add shifted digits to a15
p .x2 a18.7i
s a18.op7 A
s a18.cc7 C
s a18.rp7 1
p a18.7o .x3
p .x2 a15.2i
s a15.op2 β
s a15.cc2 0
# N4D cycle 7
# Await FT data
p D-2 a18.5i  # N4D'
s a18.op5 0
s a18.cc5 0
s a18.rp5 5
p a18.5o .x1
# N2D cycle 7
# Await FT data
p D-1 a18.5i  # N2D'
s a18.op5 0
s a18.cc5 0
s a18.rp5 5
p a18.5o .x1
# F.T. order uses E-8=3,4,5 for G-9, G-10, G-11.
p sft.4o G-9
p sft.5o G-10
p sft.6o G-11
# Read function table for F.T., using C to trigger sending the argument.
# Note this uses A+2 addressing.
p G-9 f1.2i
s f1.op2 A+2
s f1.cl2 C
s f1.rp2 1
p f1.C N-10

p G-10 f2.2i
s f2.op2 A+2
s f2.cl2 C
s f2.rp2 1
p f2.C N-10

p G-11 f3.2i
s f3.op2 A+2
s f3.cl2 0
s f3.rp2 1
p f3.C N-10

# Look up signs from the tables, too.
s f1.mpm1 T
s f1.mpm2 T
s f2.mpm1 T
s f2.mpm2 T
s f3.mpm1 T
s f3.mpm2 T
# Printer
# TODO Add some way to specify printer wiring to the simulator.  This
# instruction set expects to print a1, a2, and a15-20.
p O-1 i.Pi  # Pr.
p i.Po C-5  # fetch+decode
s pr.1 P    # a1
s pr.2 P    # a1
s pr.3 P    # a2
s pr.4 P    # a2
s pr.5 P    # a15
s pr.6 P    # a15
s pr.7 P    # a16
s pr.8 P    # a16
s pr.9 P    # a17
s pr.10 P   # a17
s pr.11 P   # a18
s pr.12 P   # a18
s pr.13 P   # a19
s pr.14 P   # a19
s pr.15 P   # a20
s pr.16 P   # a20
# Reader
p O-2 i.Ri  # Rd.
p i.Ro C-5  # fetch+decode