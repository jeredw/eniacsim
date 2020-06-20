# Reconstruction of ENIAC 60 order code
# (after Clippinger 1948: A Logical Coding System Applied to the ENIAC)
#
# The 60 order code was a theoretical instruction set for the ENIAC with 60
# instructions, planned in some detail but never built.  It was designed for
# the original ENIAC hardware with minimal extra circuitry.  The actual setup
# charts are supposedly in a box in the National Archives - this is an
# independent reconstruction based on a 1948 report available on the web, which
# I'll call [LCS48].
#
# When it comes down to it [LCS48] is vague about several instructions.  The
# 1949 report "Description and Use of the ENIAC Converter Code" [DU49]
# describes a derived instruction set in better detail and offers some clues,
# but many questions remain.
#
# Implementation notes
#
# Timings.  [LCS48] cycle timings must have been provided by the marketing
# department, as they are suspiciously low for some instructions like X.
# [DU49] has more conservative timings despite its faster 6 cycle fetch, so is
# likely more reliable.  When [LCS48] does report extra cycles, it hints at
# more complex sequences, but a low cycle count just means that the complexity
# hides behind an overlapping fetch.
#
# Clearing a13.  Several instructions require a13 to be 0 before execution and
# guarantee it to be clear after.  This is indicated as (13)0 -- (13)0 in the
# before and after columns in [LCS48] Section IV, which [DU49] corroborates
# with an explicit footnote.  This convention saves a lot of program steps.
# Reserving a13 hints at more complex sequences.
#
# Accumulator inputs.  Many instructions require shifting digits, and that must
# be done at accumulator inputs.  There are 20*5 inputs, so you wouldn't expect
# this to be a bottleneck, but the instruction set places a lot of pressure on
# a15 and a13.  Two of a13 and a15's ten inputs are tied up for the multiplier,
# and many of the rest have implied connections to support instructions that
# don't reserve a13.
#
# When a13 is reserved, it's possible to stash aN in a13, use one of aN's
# inputs, and then transfer the answer and restore aN.  The text for C.T.
# mentions using a shifter on a10 to accomplish the required PC shift this way.
#
# Reusing programs.  a15 and a13 do so much that it is vital to reuse their
# programs for multiple sequences.  This can be accomplished either with pulse
# amplifiers or dummy programs.  Pulse amplifiers are nicer because dummy
# programs introduce an extra delay cycle.
#
# F.T. addressing.  F.T. fetches data from any function table in 13 cycles by
# reusing fetch machinery without triggering decode.  But function table input
# #1 is set up to send J-4 which would trigger PC increment and advance the
# order selector ring, which is problematic because F.T. has no operand.
#
# The Axial flow listing in [LCS48] has "N4D 04 86 8l F.T." which - apart from
# the fact that FT2's constant data has address 90 in that listing - suggests
# that we're supposed to code 3-5 instead of 0-2 for function table 1/2/3 for
# F.T.  [DU49] corroborates this encoding and clarifies that instruction
# lookups use A-2 and data lookups use A+2 to permit using lines -2 through 101
# of each table.  Thus "N4D 04 86 8l F.T." would index line 88 of table 2,
# which in the A-2 instruction address space is (confusingly) line 90.
#
# If you wrote "N4D 01 00 8l F.T." by mistake, it might either hang or store an
# instruction line and skip one order, depending on the wiring.  The world may
# never know.
#
# Instruction sequences
#
# C/xl/xt.  For most cases, just
#   0. a15:0C1                          0. aL:AC1
#   1. aT:AC1, a15:β01                  1. a15:AC1, aL:α01
# 8t is a special case because a8.A=5.
#   0. a15:0C1
#   1. a8:AC1, a15:δ01
#
# X. Claims to take 15 cycles, but the multiply takes 14 and fetch takes 7.
# [DU49] claims a more honest 20 cycles with its 6 cycle fetch.
#
#   0. a12:0C1
#   1. a15:AC1, a12:α01 # Get a15 in icand. [2]
#   2. Multiply (incl. a13:AC1, a15:β01)
#  17. (answer in a15)
#
# ÷. a4 (quotient) and a9 (shifter) aren't mentioned as mutated, so assume we
# preserve those in a13 and a15.
#
#   0. a5:0C1
#   1. a15:AC1, a5:α01  # Load numerator [2]
#   2. a4:AC1, a15:β01  # Save a4 in a13 via a15 [1]
#   3. a15:AC1, a13:β01 #   (finish saving a4) [2]
#   4. a9:AC1, a15:β01  # Save a9 in a15 [1]
#   <Divide>
#  90. a9:0C1
#  91. a15:AC1, a9:α01  # Restore a9 [2]
#  92. a4:AC1, a15:β01  # Get quotient in a15 [1]
#  93. a13:AC1, a4:β01  # Restore a4 [1]
#
# √. a7 (2*root) and a9 (shifter) aren't mentioned as mutated, so assume we
# preserve those in a13 and a15.  The 5x reception requires a dedicated repeat
# program on a15 (can't reuse the normal receive program).
#
#   0. a5:0C1
#   1. a15:AC1, a5:α01  # Load numerator [2]
#   2. a7:AC1, a15:β01  # Save a7 in a13 via a15 [1]
#   3. a15:AC1, a13:β01 #   (finish saving a7) [2]
#   4. a9:AC1, a15:β01  # Save a9 in a15 [1]
#   <Square root>
#  90. a9:0C1
#  91. a15:AC1, a9:α01  # Restore a9 [2]
#  92. a7:AC5, a15:β05  # Note repeat 5: get 10*root in a15 [1]
#  97. a13:AC1, a7:β01  # Restore a7 [1]
#
# M. Complement number in accumulator 15 in 7 cycles.  [DU49] has this as 8
# cycles, assume it's just
#   0. a15:AC1, a13:β01  # Load a15. [2]
#   1. a13:SC1, a15:β01  # XXX should this be βC1? [1]
#
# DS. Drop sign of a15 in "7" add cycles.  This uses a13 and so can use a temp
# accumulator to conserve inputs, let's say
#   0. a2:AC1, a13:γ01  # Save a2 [1]
#   1. a15:AC1, a2:β01  # Receive a15 without sign [2,2+deleter]
#   2. a2:AC1, a15:β01  # Transmit answer [1]
#   3. a13:AC1, a2:γ01  # Restore a2 [1]
#
# NxD. N(x+1)D can do its thing and then reuse N(x)D programming.  Let's
# suppose a1 has the << 2 and << 4 inputs required for N4D and N6D.
#
# N6D.
#   0. a1:AC1, a13:γ01  # Save a1 [1]
#   ...
#   6. os.i, a1:β01     # Receive << 4 shifted digits at a1. [1,1<<4]
#   7. a1:AC1, a15:β01  # Add digits into a15 [1]
#   8. (to N4D+1)
# N4D.
#   0. a1:AC1, a13:γ01  # Save a1 (unless from N6D) [1]
#   ...
#   6. os.i, a1:γ01     # Receive << 2 shifted digits at a1 [1,1<<2]
#   7. a1:AC1, a15:β01  # Add digits into a15 [1]
#   8. a13:AC1, a1:δ01  # Restore a1 [1]
#   9. (to N2D)
# N2D.
#   ...
#   6. os.i, a15:β01    # Add digits into a15 [1]
#
# 6(11,10,9) and 6(8,7).
# These are listed as 7 cycles in [LCS48] but as 10 in [DU49].  They could
# overlap fetch in theory, but care would be needed to avoid conflicts because
# a6 is also the PC, e.g.
#   0. -                                      0. fetch uses a6
#   1. a6(11,10,9) += a15(11,2,1); a15 = 0    1. -
#   2. -                                      2. fetch uses a6
#   3. a13(11,2,1) += a6(11,10,9); a15 = 0    3. -
#   4. a15 += a13; a13 = 0                    4. fetch uses a6
#
# Section IV suggests this exact sequence, modulo a typo of a1 for a13.  So
# although saving through a13 would allow shifting through other accumulators,
# assume two inputs on a6 and two inputs on a13 are used to support these
# instructions.
#
# 6R3/6R6/C.T.
# 6R3/6R6 are 13 cycles and C.T. is 14 for taken and 8 for not-taken branches.
# They use a13 even though the descriptions don't mention it.  The cycle counts
# should be reliable, because the next fetch is dependent.  [LCS48] does not
# specify what happens to the unused digits of a15 for 6Rx, but [DU49]
# specifies that they're cleared, so let's go with that.
#
# 6R3 (using a4 for shifting)
# (7) 0. a4:AC1, a13:γ01  # Save a4 [1]
#     1. a6:AC1, a4:β01   # [5,5(11-4)]
#     2. a15:AC1, a4:γ01  # [2,2(3-1)]
#     3. a4:AC1, a6:δ01   # Set new PC [1]
#     4. a13:AC1, a4:δ01  # Restore a4 [1]
#     5. (dependent fetch)
#
# C.T. Yet another new stepper is snuck in to the text above [LCS48] Section IV
# ("P-M discriminator no. 2").  Without this, we'd need a separate dedicated
# accumulator for magnitude discrimination.  It's curious that C.T.-taken is 14
# cycles, since 13 cycles seems possible.  [DU49] has C.T. as 6R3+1 cycle too,
# so I must be missing something...
# (7) 0. a15:AC1, discriminate P-M
#     M1. (dependent fetch)    P1. a10:AC1, a13:γ01  # Save a10 [1]
#       									     P2. a6:AC1, a10:β01   # a6(11-7,000,6-4). [5,5*]
#       									     P3. a10:AC1, a6:δ01   # Set PC [1]
#       									     P4. a13:AC1, a10:γ01  # Restore a10 [1]
#                              P5. (dependent fetch)
#
# 18<->20. Uncharacteristically, [LCS48] lists 18<->20 as taking 9 cycles, even
# though it can be hidden behind an overlapping fetch.  Maybe marketing didn't
# like this instruction; it's gone from [DU49].  It's also possible that Sh
# wiring complicates matters for a20, e.g. limiting its connections so it
# doesn't work like other xl orders.
#
# Sh/Sh'. The vaguest instructions are Sh <jk> and Sh' <jk>.  [DU49] suggests
# these were logical, not arithmetic shifts.  Sh operates on a15 alone and Sh'
# also operates on a12' to shift a15=abcdefghij as
#
#   Shift(')    jk    New a12(')  New a15
#   abcdefghij  01 -> 000000000a  bcdefghij0
#               02 -> 00000000ab  cdefghij00
#               03 -> 0000000abc  defghij000
#               04 -> 000000abcd  efghij0000
#               05 -> 00000abcde  fghij00000
#               95 -> fghij00000  00000abcde
#               94 -> ghij000000  0000abcdef
#               93 -> hij0000000  000abcdefg
#               92 -> ij00000000  00abcdefgh
#               91 -> j000000000  0abcdefghi
#
# We're told only that this takes 20 cycles and uses a13.  Just fetching jk
# takes 13 cycles, so decoding jk and doing the work must fit in 7 cycles.
# [DU49] uses its extra code space to avoid requiring an operand and reports 9
# cycles.  That makes sense for Sh but not Sh', unless they did some really
# fancy accumulator wiring to permit a15 and a12 to be computed in parallel.
#
# The most plausible datapath is a bunch of custom shift adapters.  For
# instance say aK.β shifts << k and aK.γ shifts >> (10-k).
#   0. aK:AC1, a13:γ01, a12:0C1'  # Save aK in a13 [1]
#   1. a15:A01, aK:β01, a12:α01'  # Shift a15 in aK ('save in a12) [2,2*]
#   2. aK:AC1, a15:β01            # Store new a15 [1]
#   3'. a12:AC1, aK:γ01           # Shift old a15 again in aK [1,1*]
#   4'. aK:AC1, a12:β01           # Store new a12 [1]
#   5. a13:AC1, aK:δ01            # Restore aK from a13 [1]
#   6. (next fetch)
# Steps marked ' are done only for Sh'.  This requires three inputs each on ten
# accumulators, or one input on trunk 1 and 2 inputs per shift.  TODO: It's a
# close call as to whether there are ten accumulators with three otherwise
# unused inputs, so a few accumulators might need to be wired specially for
# Sh/Sh'.
#
# [LCS48] says nothing about decoding jk, and it was gone by [DU49].  The least
# inventive solution is to throw more hardware at the problem, e.g. a P-M
# discriminator (j) wired to two additional six-stage steppers (k) analogous to
# the FT stepper.  Wilder possibilities abound... the complement of k=1-5 could
# drive a counter whose sign says to shift one step, we could somehow reuse
# instruction decode with an additional layer of output gating, etc.  Lacking
# any more evidence, I've opted to do the simple thing.

#l testdata/selftest60

# Set up the seven add cycle fetch sequence described in II: Conversion of
# digit pulses to programming pulses.  The first cycle decodes the hundreds
# digit of the PC (a6) to select one of three FT; then the FT read takes five
# cycles, and the final cycle decodes the order.
# 
# Parts of this sequence are reused to fetch arguments for N4D, N6D so are
# separately controllable.  e.g. N4D triggers D-6->J-2,E-1 but not C-5, so it
# will read from the function table but not decode a new instruction.
#
#     F.T.    ORDER    10 STAGE   MASTER   ACC 6   ACC 20   FT I  FT II  FT III  SC?
#     SEL     SEL      STEPPER    PROG.
# -----------------------------------------------------------------------------------------
# 0   E-1 i#1                              J-2     C-5
#     E-8 di                                A01     006
#     1->E-11                                E-9     |
#     2->F-1                                         |
#     3->F-2                                         |
# ---------------------------------------------------|-------------------------------------
# 1   E-9 i#2?                                       |      E-11  F-1    F-2
#                                                    |      NC|   NC|    NC|
#                                                    |       J-4   J-4    J-4
# ---------------------------------------------------|--------|-----|------|---------------
# 2                                        J-4       |        |     |      |
#                                           A01      |        |     |      |
#                                            F-3     |        |     |      |
# ---------------------------------------------------|--------|-----|------|---------------
# 3           F-3                                    |        |     |      |     F-3
#              CPP                                   |        |     |      |      001
#               F-4                                  |        |     |      |       F-5
# ---------------------------------------------------|--------|-----|----------------------
# 4                    F-5        F-5      F-4       |        |     |      |
#                       cdi        cdi      εC1      |        |     |      |
#                                          F-5       |        |     |      |
#                                           002      |        |     |      |
#                                            |       |        |     |      |
# -------------------------------------------|-------|--------|-----|------|---------------
# 5                                          |       |        |     |      |
#                                            |       |        |     |      |
#                                            |       |        |     |      |
#                                            |       |        |     |      |
# -------------------------------------------v-------v--------v-----v------v---------------
# 6           G-8      E-10                  G-8     E-10
#              i        i
# -----------------------------------------------------------------------------------------

# FT selector
# E-8 selects the third digit of a6 or a8 from data trunk 5.
p 5 ad.dp.1.3
p ad.dp.1.3 E-8
# Decode E-8=0,1,2 into E-11, F-1, or F-2.
p E-1 sft.i
p E-8 sft.di
p sft.1o E-11
p sft.2o F-1
p sft.3o F-2
# F.T. order uses E-8=3,4,5 for G-3, G-4, G-5.
p sft.4o G-3
p sft.5o G-4
p sft.6o G-5
# Clear FT selector for next fetch.
p E-9 sft.cdi

# Order selector
p os.A 3
p os.B 4
p os.o 1
# Step the instruction ring counter.
p F-3 os.Ri
p os.Ro F-4  # Used to increment PC
p G-8 os.i

# Ten stage stepper
# F-6 selects the tens digit of the order from data trunk 1.
p 1 ad.dp.2.2
p ad.dp.2.2 F-6
# Clear prior to decode
p F-5 st.cdi
# Decode tens digit and trigger corresponding MP stepper input.
p E-10 st.i
p F-6 st.di
p st.1o p.Ai
p st.2o p.Bi
p st.3o p.Ci
p st.4o p.Di
p st.5o p.Ei
p st.6o p.Fi
p st.7o p.Gi
p st.8o p.Hi
p st.9o p.Ji
p st.10o p.Ki

# Master programmer (input side, see below for stepper outputs)
# F-7 selects the ones digit of the order from data trunk 1.
p 1 ad.dp.3.1
p ad.dp.3.1 F-7
# Clear prior to decode
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
# Decode ones digit.  (Tens decode triggers the appropriate stepper.)
p F-7 p.Adi
p F-7 p.Bdi
p F-7 p.Cdi
p F-7 p.Ddi
p F-7 p.Edi
p F-7 p.Fdi
p F-7 p.Gdi
p F-7 p.Hdi
p F-7 p.Jdi
p F-7 p.Kdi
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
# Configure decade switches so decades never overflow.
s p.d1s1 9
s p.d2s1 9
s p.d3s1 9
s p.d4s1 9
s p.d5s1 9
s p.d6s1 9
s p.d7s1 9
s p.d8s1 9
s p.d9s1 9
s p.d10s1 9
s p.d11s1 9
s p.d12s1 9
s p.d13s1 9
s p.d14s1 9
s p.d15s1 9
s p.d16s1 9
s p.d17s1 9
s p.d18s1 9
s p.d19s1 9
s p.d20s1 9

# Accumulator 6 (program counter)
p a6.A 5
p 2 a6.α
p 1 a6.β
p ad.permute.10 a6.γ  # 2 (11,2,1,0s)
p ad.permute.12 a6.δ  # 2 (0,0,0,2,1,0s)
# NB ε is the dummy for PC increment
# Transmit value for FT selection.
p J-2 a6.5i
s a6.op5 A
s a6.cc5 0
s a6.rp5 1
p a6.5o E-9  # Clear FT selector
# Transmit value for FT argument.
p J-4 a6.6i
s a6.op6 A
s a6.cc6 0
s a6.rp6 1
p a6.6o F-3
# PC increment
p F-4 a6.1i
p a6.op1 ε
p a6.cc1 C
# Dummy to trigger order selector
p F-5 a6.7i
p a6.op7 0
p a6.cc7 0
p a6.rp7 2
p a6.7o G-8
# 6l: clear and then receive from a15
#p E-4 a6.8i
#s a6.op8 0
#s a6.cc8 C
#s a6.rp8 1
#p a6.8o C-4
#p C-4 a6.9i
#s a6.op9 α
#s a6.cc9 0
#s a6.rp9 1
#p a6.9o C-5  # Re-trigger fetch sequence
# 6R3 and 6R6
p R-1 a6.8i  # 6R3 +1: a10 += a6(11-4); a6 = 0
s a6.op8 A
s a6.cc8 C
s a6.rp8 1
p R-3 a6.4i  # 6R3 +3: a6 += a10; a10 = 0
s a6.op4 β
# 6(11,10,9)
p T-1 a6.10i # 6(11,10,9) +1: a6(11,10,9) += a15(11,2,1); a15 = 0
s a6.op10 γ
s a6.cc10 0
s a6.rp10 1
p a6.10o T-2 # 6(11,10,9) +2
p T-3 a6.2i  # 6(11,10,9) +3: a13(11,2,1) += a6(11,10,9)
s a6.op2 A
# 6(8,7)
p T-7 a6.11i # 6(8,7) +1: a6(8,7) += a15(2,1); a15 = 0
s a6.op11 δ
s a6.cc11 0
s a6.rp11 1
p a6.11o T-8 # 6(8,7) +2
p T-9 a6.3i  # 6(8,7) +3: a13(11,2,1) += a6(11,10,9)
s a6.op3 A

# Function tables
p 1 f1.arg
p f1.A 3
p f1.B 4
p E-11 f1.1i # Read instruction or operand.
s f1.op1 A-2
s f1.cl1 NC
s f1.rp1 1
p f1.NC J-4  # Send argument for instruction or operand.
p G-3 f1.2i  # Read F.T. data
s f1.op2 A+2
s f1.cl2 C
s f1.rp2 1
p f1.C N-10  # Send argument for F.T.
s f1.mpm1 T
s f1.mpm2 T

p 1 f2.arg
p f2.A 3
p f2.B 4
p F-1 f2.1i  # Read instruction or operand.
s f2.op1 A-2
s f2.cl1 NC
s f2.rp1 1
p f2.NC J-4  # Send argument for instruction or operand.
p G-4 f2.2i  # Read F.T. data
s f2.op2 A+2
s f2.cl2 C
s f2.rp2 1
p f2.C N-10  # Send argument for F.T.
s f2.mpm1 T
s f2.mpm2 T

p 1 f3.arg
p f3.A 3
p f3.B 4
p F-2 f3.1i  # Read instruction or operand.
s f3.op1 A-2
s f3.cl1 NC
s f3.rp1 1
p f3.NC J-4  # Send argument for instruction or operand.
p G-5 f3.2i  # Read F.T. data
s f3.op2 A+2
s f3.cl2 0
s f3.rp2 1
p f3.C N-10  # Send argument for F.T.
s f3.mpm1 T
s f3.mpm2 T

# Decode opcodes to program lines following Table 2.I.
#
# The report describes two slightly different codes, a regular code and an
# alternate code with seven orders reconfigured.  Alternate orders are
# annotated below as <N>a where N is the opcode.
#
# The listings in the report use the regular code, so that is wired up here.
# In some cases wiring for the alternate code is present but commented out.
p p.A1o C-1   # C   (00)
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
p p.D1o E-5   # X   (30)  15 cycles
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
p p.F6o E-3   # ÷   (55)  ~75 cycles
p p.G1o B-1   # √   (60)  ~75 cycles
p p.G2o O-1   # Pr. (61)  60 cards / min
p p.G3o O-2   # Rd. (62)  100 cards / min
p p.G4o V-3   # M   (63)
p p.G5o H-8   # 19t (64)
p p.G6o H-10  # DS  (65)
p p.H1o V-5   # N2D (70)  14 cycles
#p p.H2o      # Halt (71) Doesn't decode to anything, so stops
p p.H3o H-11  # F.T.(72)  13 cycles
p p.H4o E-6   # 6R3 (73)  13 cycles
#p p.H4o E-4  # 6l  (73a) 9 cycles
p p.H5o E-7   # 6R6 (74)  13 cycles
#p p.H5o V-1  # 6t  (74a)
p p.H6o G-2   # C.T.(75)  14/8 cycles
p p.J1o D-7   # AB  (80)
p p.J2o D-8   # CD  (81)
p p.J3o D-9   # EF  (82)
p p.J4o D-10  # GH  (83)
p p.J5o D-11  # JK  (84)
p p.J6o C-10  # Sh' (85)  20 cycles
p p.K1o C-11  # Sh  (90)  20 cycles
p p.K2o C-3   # 18<->20 (91)  9 cycles
#p p.K2o C-8  # 20l (91a)
p p.K3o V-2   # 6(11,10,9) (92)
# Table 2.I has "H-0" for order 92a. Assume this is a typo for H-9 which is the
# only H-line missing, and makes sense from pulse amp grouping.
#p p.K3o H-9  # 20t        (92a)
p p.K4o V-4   # 6(8,7)     (93)
#p p.K4o V-8  # N3D8       (93a) 20 cycles
p p.K5o V-6   # N4D        (94)  20 cycles
#p p.K5o ???  # N3D6       (94a) 20 cycles
p p.K6o V-7   # N6D        (95)  26 cycles
#p p.K6o ???  # N6D6       (95a) 26 cycles

# Use a tree of pulse amplifiers to combine order lines into common control
# signals as described in Table 2.II.
#
# The most important of these signals are C-5 and its duals D-6, J-2, E-1,
# which trigger the next order fetch overlapped with the current order
# execution.  (Longer duration orders and control transfers must have separate
# control wiring and not feed into C-5.)
#
# Using pulse amplifiers is crucial to conserve program inputs.  Several
# apparently unused groupings have been commented out to make room.

# Box 1: Listen orders (xl) -> C-1
# Note that C-1 is also the encoding for 00 (C) which clears Acc 15.
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
#p pa.2.sa.6  C-8  # 20l
#p pa.2.sb.6  C-1

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
#p pa.4.sa.1  H-9  # 20t
#p pa.4.sb.1  D-3

# Box 3: Constant transfers (uv) -> D-4
p pa.4.sa.2  D-7  # AB
p pa.4.sb.2  D-4
p pa.4.sa.3  D-8  # CD
p pa.4.sb.3  D-4
p pa.4.sa.4  D-9  # EF
p pa.4.sb.4  D-4
p pa.4.sa.5  D-10 # GH
p pa.4.sb.5  D-4
p pa.4.sa.6  D-11 # JK
p pa.4.sb.6  D-4

# Box 4: Misc remaining 7 cycle ops -> D-5
# TODO: Why are these grouped in box 4 instead of listed separately?
p pa.4.sa.7 V-2  # 6(11,10,9)
p pa.4.sb.7 D-5
p pa.4.sa.8 V-3  # M
p pa.4.sb.8 D-5
p pa.4.sa.9 V-4  # 6(8,7)
p pa.4.sb.9 D-5

# Talk or constant transfers -> B-3
p pa.4.sa.10 D-3  # xt
p pa.4.sb.10 B-3
p pa.4.sa.11 D-4  # uv
p pa.4.sb.11 B-3
# Dual of constant transfers -> J-3
p pa.5.sa.1 D-4  # uv
p pa.5.sb.1 J-3
# Dual of (misc) -> B-4
p pa.5.sa.2 D-5  # misc
p pa.5.sb.2 B-4

# C-5 triggers fetching _and decoding_ another instruction.
# C-1 (xl), D-3 (xt), D-4 (uv), D-5 (misc) as well as C-2 (8t), V-1 (6t in
# alternate code), and H-10 (DS) all do so immediately upon decode.
p pa.5.sa.3 C-1  # xl
p pa.5.sb.3 C-5
p pa.5.sa.4 D-3  # xt
p pa.5.sb.4 C-5
p pa.5.sa.5 D-4  # uv
p pa.5.sb.5 C-5
p pa.5.sa.6 D-5  # misc
p pa.5.sb.6 C-5
p pa.5.sa.7 C-2  # 8t
p pa.5.sb.7 C-5
#p pa.5.sa.8 V-1  # 6t
#p pa.5.sb.8 C-5
p pa.5.sa.9 H-10 # DS
p pa.5.sb.9 C-5

# Not documented what this does but assume it's used for 6l.
#p pa.5.sa.10 C-4  # Cycle 2 of 6l, from a6
#p pa.5.sb.10 J-1  # Trigger a15 to send, part of xl.

#p pa.5.sa.11 V-5  # N2D
#p pa.5.sb.11 D-1
p pa.6.sa.1 V-5  # N2D
p pa.6.sb.1 D-6
p pa.6.sa.2 V-6  # N4D
p pa.6.sb.2 D-6
p pa.6.sa.3 V-7  # N6D
p pa.6.sb.3 D-6
#p pa.6.sa.4 V-8  # N3D8
#p pa.6.sb.4 C-6
#p pa.6.sa.5 V-6  # N4D
#p pa.6.sb.5 D-2

# Note C-11 is also the encoding for Sh.
p pa.6.sa.6 C-10 # Sh'
p pa.6.sb.6 C-11 # Sh

# D-6 controls operand fetch specifically, separately from decode.
#p pa.6.sa.7 D-1  # dual of N2D
#p pa.6.sa.7 V-5  # N2D
#p pa.6.sa.8 C-6  # dual of N*D*
#p pa.6.sb.8 D-6
#p pa.6.sa.9 D-2  # dual of N4D
#p pa.6.sa.9 V-6  # N4D
p pa.6.sb.9 D-6
p pa.6.sa.10 C-5  # fetch+decode -> decode
p pa.6.sb.10 D-6
p pa.6.sa.11 C-11 # Sh/Sh' take a shift amount
p pa.6.sb.11 D-6

# J-2 and E-1 are duals of D-6
p pa.7.sa.1 D-6
p pa.7.sb.1 J-2  # PC update sequence
p pa.7.sa.2 D-6
p pa.7.sb.2 E-1  # FT read sequence

#p pa.7.sa.3 C-8  # 20l
#p pa.7.sb.3 K-11

# The pulse amplifier wiring below isn't specified in the report, but is
# convenient to this reconstruction.

p pa.7.sa.3 L-5  # 13t
p pa.7.sb.3 T-5  # a13 send/clear

# 6t/8t, i.e. xt orders that use digit trunk 5
#p pa.7.sa.4 V-1  # 6t
#p pa.7.sb.4 N-7  # a15 receive
p pa.7.sa.5 C-2  # 8t
p pa.7.sb.5 N-7  # a15 receive

# M needs to send from a15, read at a13
p pa.7.sa.6 V-3  # M +0
p pa.7.sb.6 J-1  # a15 send/clear
p pa.7.sa.7 N-1  # M +1
p pa.7.sb.7 B-3  # a15 receive

# F.T. order should trigger an FT read
p pa.7.sa.8 H-11 # F.T.
p pa.7.sb.8 E-1  # ft selector
p pa.7.sa.9 H-11 # F.T.
p pa.7.sb.9 J-1  # a15 send/clear (used for clear)

# N2D
p pa.7.sa.10 N-6  # N2D +6
p pa.7.sb.10 B-3  # a15 receive
p pa.7.sa.11 N-6  # N2D +6
p pa.7.sb.11 C-5  # fetch+decode (in same cycle as data written back)

# DS sends from a15 then reads from a13
p pa.8.sa.1 H-10  # DS +0
p pa.8.sb.1 J-1   # a15 send/clear
p pa.8.sa.2 N-9   # DS +2
p pa.8.sb.2 B-3   # a15 receive

# 18<->20
p pa.8.sa.3 C-3   # 18<->20 +0
p pa.8.sb.3 J-1   # a15 send/clear (just clear)
p pa.8.sa.4 C-3   # trigger fetch+decode immediately
p pa.8.sb.4 C-5   # fetch+decode
p pa.8.sa.5 N-11  # 18<->20 +1
p pa.8.sb.5 B-3   # a15 receive
p pa.8.sa.6 M-2   # 18<->20 +3
p pa.8.sb.6 J-1   # a15 send/clear

# 6(11,10,9)
p pa.8.sa.7 T-1   # 6(11,10,9) +1
p pa.8.sb.7 J-1   # a15 send/clear
p pa.8.sa.8 T-4   # 6(11,10,9) +4
p pa.8.sb.8 B-3   # a15 receive
p pa.8.sa.9 T-4   # 6(11,10,9) +4
p pa.8.sb.9 T-5   # a13 send/clear

# 6(8,7)
p pa.8.sa.10 T-7  # 6(8,7) +1
p pa.8.sb.10 J-1  # a15 send/clear
p pa.8.sa.11 T-10 # 6(8,7) +4
p pa.8.sb.11 B-3  # a15 receive
p pa.5.sa.11 T-10 # 6(8,7) +4
p pa.5.sb.11 T-5  # a13 send/clear

# 6R3
p pa.4.sa.1 R-2   # 6R3 +2
p pa.4.sb.1 J-1   # a15 send/clear
p pa.2.sa.6 R-4   # 6R3 +4
p pa.2.sb.6 T-5   # a13 send/clear

# X
p pa.5.sa.8 M-9   # X +1
p pa.5.sb.8 J-1   # a15 send/clear

#p pa.5.sa.10
#p pa.5.sb.10
#p pa.5.sa.11
#p pa.5.sb.11
#p pa.6.sa.4
#p pa.6.sb.4
#p pa.6.sa.5
#p pa.6.sb.5
#p pa.6.sa.7
#p pa.6.sa.7
#p pa.6.sa.8
#p pa.6.sb.8
#p pa.6.sa.9
#p pa.6.sa.9
#p pa.7.sa.4
#p pa.7.sb.4

# Digit trunks
#
# The report mentions that for 1l, a1.α=2 and a15.A=2; for AB, figure 3.2 has
# c.o=1, a11.δ=1, and a15.β=1 (with a note that says "see digit tray hook-up".)
# This last assignment overlaps with the canonical multiplier correction wiring
# from the Technical Manual which has the correction term on a15.β, but it's
# plausible that's shared with trunk 1.
# 
# Since a15 is the multiplier RHPP, α is dedicated for that.  β is wired to 1.
# a15 has to be able to receive shifted function table data for F.T., so that's
# γ.  Assuming the function table argument is on a separate trunk 5 to simplify
# fetch/execute overlap, a8 must transmit on that trunk for F.T.; then to
# support 8t, a15.δ=5 (6t timing also supports this).
#
# a6.α   trunk 2
# a6.β   for 6(11,10,9)
# a6.γ   for 6(8,7)
# a6.δ   trunk 1 for 6Rx and C.T.
# a6.ε   dummy for increment
#
# a15.α  dedicated for multiplier RHPP
# a15.β  1, from report; also correction term
# a15.γ  shifted data from function table (for F.T.)
# a15.δ  5, to support 6t/8t and 6(...) ops
# a15.ε
#
# a13.α  dedicated for multiplier LHPP
# a13.β  correction term; also 2?  can't be 1
# a13.γ  trunk 1 so a13 can be a temporary
# a13.δ  for 6(11,10,9)
# a13.ε  for 6(8,7)
#
# Why not just have a15.A=1 and aN.α=1, using a single trunk for accumulator
# transfers?
#
#                quo num  PC den     shi     ier ica lhp     rhp 
#   | a1| a2| a3| a4| a5| a6| a7| a8| a9|a10|a11|a12|a13|a14|a15|a16|a17|a18|a19|a20|
#  -+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
#  α| 2 | 2 | 2 | A | 2 | 2 | 2 | 2 |*11| 2 | 2 | 2 | 6 | 2 | 7 | 2 | 2 | 2 | 2 | 2 |
#  β|   |   |   | 1 |   |*  | 1 |   | 2 |   |*3 |   |8=2|   |9=1|   |   |   |   |   |
#  γ|   |   |   |   |B=1|*  | A |   |   |   |   |   | 1 |   |*4 |   |   |   |   |   |
#  δ|   |   |   |   |   |   |   |   |   |   | 1 |   |   |   | 5 |   |   |   |   |   |
#  ε|   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |   |
#  -+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
#  A| 1 | 1 | 1 | 1 |B=1| 5 |B=1| 5 |B=1| 1 | 1 | 1 | 1 | 1 | 2 | 1 | 1 | 1 | 1 | 1 |
#  S|   |   |   |   |   |   |B=1|   |   |   | 8 | 9 | 1 |   |   |   |   |   |   |   |
#
#  1  - most accumulators transmit on 1
#  2  - most accumulators receive on 2
#  3  - ft data A
#  4  - ft data B
#  5  - ft argument
#  6,7- multiplier partial product digits (exclusive)
#  8=2- multiplier correction terms (shared)
p 8 2
#  9=1- multiplier correction terms (shared)
p 9 1
#  A  - divider/square rooter answer (exclusive)
#  B=1- divider/square rooter shift (shared)
p 11 1

# Special digit adapters

# Shift for N4D.
p 1 ad.permute.4
s ad.permute.4 0,0,0,0,0,0,0,2,1,0,0
# Shift for N6D.
p 1 ad.permute.5
s ad.permute.5 0,0,0,0,0,2,1,0,0,0,0

# F.T. shifts data to the left.
p 3 ad.permute.6
s ad.permute.6 11,6,5,4,3,2,1,0,0,0,0
p 4 ad.permute.7
s ad.permute.7 11,6,5,4,3,2,1,0,0,0,0

# Delete sign for DS
p 2 ad.permute.8
s ad.permute.8 0,10,9,8,7,6,5,4,3,2,1

# For 6(11,10,9).
p 2 ad.permute.10
s ad.permute.10 11,2,1,0,0,0,0,0,0,0,0
p 5 ad.permute.11
s ad.permute.11 11,0,0,0,0,0,0,0,0,10,9

# For 6(8,7).
p 2 ad.permute.12
s ad.permute.12 0,0,0,2,1,0,0,0,0,0,0
p 5 ad.permute.13
s ad.permute.13 0,0,0,0,0,0,0,0,0,8,7

# For 6R3.
p 5 ad.permute.14
s ad.permute.14 11,10,9,8,7,6,5,4,0,0,0
p 2 ad.permute.15
s ad.permute.15 0,0,0,0,0,0,0,0,3,2,1

# For 6R6.
p 5 ad.permute.16
s ad.permute.16 11,10,9,8,7,0,0,0,0,0,0
p 2 ad.permute.17
s ad.permute.17 0,0,0,0,0,6,5,4,3,2,1

# Accumulator 1
p a1.A 1
p 2 a1.α
# 1l: Clear and then receive.
p S-1 a1.5i
s a1.op5 A  # NB this is AC1 in Figure 3.3 not 0C1.
s a1.cc5 C
s a1.rp5 1
p a1.5o O-8
p O-8 a1.1i
s a1.op1 α
# 1t: Transmit and hold.
p V-9 a1.2i
s a1.op2 A

# Accumulator 2
p a2.A 1
p 2 a2.α
# 2l: Clear and then receive.
p S-2 a2.5i
s a2.op5 A
s a2.cc5 C
s a2.rp5 1
p a2.5o a2.1i
s a2.op1 α
# 2t: Transmit and hold.
p S-7 a2.2i
s a2.op2 A

# Accumulator 3
p a3.A 1
p 2 a3.α
# 3l: Clear and then receive.
p S-3 a3.5i
s a3.op5 A
s a3.cc5 C
s a3.rp5 1
p a3.5o a3.1i
s a3.op1 α
# 3t: Transmit and hold.
p S-8 a3.2i
s a3.op2 A

# Accumulator 4 (quotient)
# a4.A 1
p 10 a4.α  # d.ans
p 2 a4.β
p 1 a4.γ
# 4l: Clear and then receive.
p S-4 a4.5i
s a4.op5 A
s a4.cc5 C
s a4.rp5 1
p a4.5o a4.1i
s a4.op1 β
# 4t: Transmit and hold.
p S-9 a4.2i
s a4.op2 A

# Accumulator 5 (numerator)
p a5.A 11
p 2 a5.α
p 11 a5.γ  # divider shift
# 5l: Clear and then receive.
p S-5 a5.5i
s a5.op5 A
s a5.cc5 C
s a5.rp5 1
p a5.5o a5.1i
s a5.op1 α
# 5t: Transmit and hold.
p S-10 a5.2i
s a5.op2 A
# 

# Accumulator 6 wiring is above in fetch section.

# Accumulator 7 (denominator)
p a7.A 11
p a7.S 11
p 2 a7.α
p 1 a7.β
p 10 a7.γ
# 7l: Clear and then receive.
p S-6 a7.5i
s a7.op5 A
s a7.cc5 C
s a7.rp5 1
p a7.5o a7.1i
s a7.op1 α
# 7t: Transmit and hold.
p S-11 a7.2i
s a7.op2 A

# Accumulator 8
p a8.A 5  # for F.T.
p 2 a8.α
# 8l: Clear and then receive.
p L-7 a8.5i
s a8.op5 A
s a8.cc5 C
s a8.rp5 1
p a8.5o a8.1i
s a8.op1 α
# 8t: Transmit and hold.
p C-2 a8.2i
s a8.op2 A
# F.T.: a8 is the address.  It is automatically post-incremented.
p H-11 a8.3i  # F.T. +0: decode ft number
s a8.op3 A
p N-10 a8.6i  # F.T. +2: send argument to ft
s a8.op6 A
s a8.cc6 0
s a8.rp6 1
p a8.6o a8.7i # F.T. +3: increment ft address
s a8.op7 ε
s a8.cc7 C
s a8.rp7 1
p a8.7o E-9   # Clear FT selector

# Accumulator 9 (shift)
p a9.A 11
p 11 ad.s.1.1
p ad.s.1.1 a9.α # shifter
p 2 a9.β
# 9l: Clear and then receive.
p L-8 a9.5i
s a9.op5 A
s a9.cc5 C
s a9.rp5 1
p a9.5o a9.1i
s a9.op1 β
# 9t: Transmit and hold.
p L-1 a9.2i
s a9.op2 A

# Accumulator 10
p a10.A 1
p 2 a10.α
p 1 a10.β
p ad.permute.14 a10.γ  # 5 (11-4)
p ad.permute.15 a10.δ  # 2 (3-1)
# 10l: Clear and then receive.
p L-9 a10.5i
s a10.op5 A
s a10.cc5 C
s a10.rp5 1
p a10.5o a10.1i
s a10.op1 α
# 10t: Transmit and hold.
p L-2 a10.2i
s a10.op2 A
# 6R3
p E-6 a10.6i  # 6R3 +0: a13 += a10; a10 = 0
s a10.op6 A
s a10.cc6 C
s a10.rp6 1
p a10.6o R-1
p R-1 a10.7i  # 6R3 +1: a10 += a6(11-4); a6 = 0
s a10.op7 γ
s a10.cc7 0
s a10.rp7 1
p a10.7o R-2
p R-2 a10.8i  # 6R3 +2: a10 += a15(3-1); a15 = 0
s a10.op8 δ
s a10.cc8 0
s a10.rp8 1
p a10.8o R-3
p R-3 a10.9i  # 6R3 +3: a6 += a10; a10 = 0
s a10.op9 A
s a10.cc9 C
s a10.rp9 1
p a10.9o R-4
p R-4 a10.3i  # 6R3 +4: a10 += a13; a13 = 0
s a10.3i β

# Accumulator 11 (ier)
p a11.A 1
p a11.S 8  # lhpp correction term
p 2 a11.α
p 1 a11.β            # from c.o
p ad.permute.6 a11.γ # from ft.A, shifted
# For constant transfer, clear 5i then receive digits 1i.
p J-3 a11.5i  # uv
s a11.op5 0
s a11.cc5 C
s a11.rp5 1
p a11.5o T-6
p T-6 a11.1i
s a11.op1 β
# 11l: Clear and then receive.
p L-10 a11.6i
s a11.op6 A
s a11.cc6 C
s a11.rp6 1
p a11.6o a11.2i
s a11.op2 α
# 11t: Transmit and hold.
p L-3 a11.3i
s a11.op3 A
# F.T.: Clear and then receive.
p H-11 a11.7i   # Wait for ft data and clear.
s a11.op7 0
s a11.cc7 C
s a11.rp7 6
p a11.7o a11.8i # Read ft data.
s a11.op8 γ
s a11.cc8 0
s a11.rp8 1
p M-10 a11.4i
s a11.op4 S

# Accumulator 12 (icand)
p a12.A 1
p a12.S 9  # rhpp correction term
p 2 a12.α
# 12l: Clear and then receive.
p L-11 a12.5i
s a12.op5 A
s a12.cc5 C
s a12.rp5 1
p a12.5o a12.1i
s a12.op1 α
# 12t: Transmit and hold.
p L-4 a12.2i
s a12.op2 A
# Multiply sequence.
p E-5 a12.6i  # X +0: a12 = 0
s a12.op6 0
s a12.cc6 C
s a12.rp6 1
p a12.6o M-9
p M-9 a12.7i  # X +1: a12 += a15; a15 = 0
p a12.op7 α
p a12.cc7 0
p a12.rp7 1
p a12.7o M-8  # X +2
# Multiplier correction term.
p M-11 a12.3i
s a12.op3 S

# Accumulator 13 (scratch/zero register, LHPP for multiply)
p a13.A 1
p a13.S 1  # used for M
p 6 a13.α  # lhpp
p 8 a13.β  # lhpp correction term (note 8=2)
p ad.permute.11 a13.γ # 5 (11,0s,10,9)
p ad.permute.13 a13.δ # 5 (0s,8,7)
p ad.permute.8  a13.ε # 2 (0,10-1)
# XXX find an input
#p 1 a13.β

# X: Sign correction
p M-10 a13.1i
s a13.op1 β
# X: Transmit partial product to RHPP and clear.
p M-7 a13.2i
s a13.op2 A
s a13.cc2 C

# 13l: Uniquely, 13l does not clear before receiving.
#p N-8 a13.1i
#s a13.op1 β
# 13t: Uniquely, 13t clears after sending.
p T-5 a13.3i
s a13.op3 A
s a13.cc3 C

# M: Receive a15, then transmit complement
# XXX find an input
#p V-3 a13.5i  # M +0: Receive a15
#s a13.op5 β
#s a13.cc5 0
##s a13.rp5 1
#p a13.5o N-1
#p N-1 a13.3i  # M +1: Transmit -a15 and clear.
#s a13.op3 S
#s a13.cc3 C
# DS: Receive a15 w/o sign, then transmit.
p H-10 a13.6i # DS +0: Receive a15 w/o sign
s a13.op6 ε
s a13.cc6 0
s a13.rp6 1
p a13.6o N-9
p N-9 a13.4i  # DS +1: Transmit a15 (w/o sign) and clear.
s a13.op4 A
s a13.op4 C
# 6(11,10,9): Receive sum and send/clear to a15
p T-3 a13.7i # 6(11,10,9) +3: a13(11,2,1) += a6(11,9,8)
s a13.op7 γ
s a13.cc7 0
s a13.rp7 1
p a13.7o T-4
# 6(8,7): Receive sum and send/clear to a15
p T-9 a13.8i # 6(8,7) +3: a13(2,1) += a6(8,7)
s a13.op8 δ
s a13.cc8 0
s a13.rp8 1
p a13.8o T-10 # 6(8,7) +4
# 6R3: Receive a10 value.
# XXX find an input
#p E-6 a13.9i  # 6R3 +0: a13 += a10; a10 = 0
#s a13.op9 β
#s a13.cc9 0
#s a13.rp9 1

# Accumulator 14
p a14.A 1
p 2 a14.α
# 14l: Clear and then receive.
p H-1 a14.5i
s a14.op5 A
s a14.cc5 C
s a14.rp5 1
p a14.5o a14.1i
s a14.op1 α
# 14t: Transmit and hold.
p L-6 a14.2i
s a14.op2 A

# Accumulator 15 (RHPP)
p a15.A 2
p 7 a15.α             # rhpp must be on α
p 9 a15.β             # rhpp correction term (note 9=1)
p ad.permute.
p 5 a15.γ             # from a8 and a6

#XXX find an input
#p ad.permute.7 a15.α  # 4 (11,6-1,0s)
#p 1 a15.β             # from aN and os.

# X: Correction term 
p M-11 a15.1i
s a15.op1 β
# X: Combine partial product from LHPP
p M-7 a15.4i
s a15.op4 β

# xt, uv, M, DS, F.T., 18<->20: Receive on β
#p B-3 a15.1i
#s a15.op1 β
# 6t/8t: Receive on γ
p N-7 a15.2i
s a15.op2 γ
# xl, F.T., M, 6(..): Send and clear.
p J-1 a15.3i
s a15.op3 A
s a15.cc3 C
# F.T.: Clear and then receive.
# XXX find an input
#p N-3 a15.5i  # F.T. +6: Read ft data.
#s a15.op5 α
#s a15.cc5 0
#s a15.5o C-5  # Trigger next instruction fetch+decode.

# N6D: Wait for operand, then receive.
p N-4 a15.6i  # N6D cycle 6: Add in shifted digits.
s a15.op6 ε
s a15.cc6 0
s a15.rp6 1
p a15.6o V-6  # Trigger N4D op for next two digits.
# N4D: Wait for operand, then receive.
p N-5 a15.7i  # N4D cycle 6: Add in shifted digits.
s a15.op7 δ
s a15.cc7 0
s a15.rp7 1
p a15.7o V-5  # Trigger N2D op for next two digits.

# Accumulator 16
p a16.A 1
p 2 a16.α
# 16l: Clear and then receive.
p H-2 a16.5i
s a16.op5 A
s a16.cc5 C
s a16.rp5 1
p a16.5o a16.1i
s a16.op1 α
# 16t: Transmit and hold.
p C-9 a16.2i
s a16.op2 A
# Dummy programs to wait for operands.
p H-11 a16.6i # F.T.
s a16.op6 0
s a16.cc6 0
s a16.rp6 6
p a16.6o N-3
p V-7 a16.7i  # N6D
s a16.op7 0
s a16.cc7 0
s a16.rp7 6
p a16.7o N-4
p V-6 a16.8i  # N4D
s a16.op8 0
s a16.cc8 0
s a16.rp8 6
p a16.8o N-5
p V-5 a16.9i  # N2D
s a16.op9 0
s a16.cc9 0
s a16.rp9 6
p a16.9o N-6  # triggers B-3 via pulse amp

# Accumulator 17
p a17.A 1
p 2 a17.α
# 17l: Clear and then receive.
p H-3 a17.5i
s a17.op5 A
s a17.cc5 C
s a17.rp5 1
p a17.5o a17.1i
s a17.op1 α
# 17t: Transmit and hold.
p H-6 a17.2i
s a17.op2 A

# Accumulator 18
p a18.A 1
p 2 a18.α
# 18l: Clear and then receive.
p H-4 a18.5i
s a18.op5 A
s a18.cc5 C
s a18.rp5 1
p a18.5o a18.1i
s a18.op1 α
# 18t: Transmit and hold.
p H-7 a18.2i
s a18.op2 A
# 18<->20
p M-1 a18.6i  # 18<->20 +2: a20 += a18; a18 = 0
s a18.op6 A
s a18.cc6 C
s a18.rp6 1
p M-2 a18.7i  # 18<->20 +3: a18 += a15; a15 = 0
s a18.op7 α
s a18.cc7 0
s a18.rp7 1

# Accumulator 19
p a19.A 1
p 2 a19.α
# 19l: Clear and then receive.
p H-5 a19.5i
s a19.op5 A
s a19.cc5 C
s a19.rp5 1
p a19.5o a19.1i
s a19.op1 α
# 19t: Transmit and hold.
p H-8 a19.2i
s a19.op2 A
# Dummies for 6(8,7)
p V-4 a19.9i  # 6(8,7) +0
s a19.op9 0
s a19.cc9 0
s a19.rp9 1
p a19.9o T-7  # 6(8,7) +1
p T-8 a19.10i # 6(8,7) +2
s a19.op9 0
s a19.cc9 0
s a19.rp9 1
p a19.10o T-9 # 6(8,7) +3
# Dummies for 6(11,10,9)
p V-2 a19.11i # 6(11,10,9) +0
s a19.op11 0
s a19.cc11 0
s a19.rp11 1
p a19.11o T-1 # 6(11,10,9) +1
p T-2 a19.12i # 6(11,10,9) +2
s a19.op12 0
s a19.cc12 0
s a19.rp12 1
p a19.12o T-3 # 6(11,10,9) +3

# Accumulator 20
p a20.A 1
p 2 a20.α
p 1 a20.β
# 20l (alternate set): Clear and then receive.
#p C-8 a20.5i
#s a20.op5 A
#s a20.cc5 C
#s a20.rp5 1
#p a20.5o a20.1i
#s a20.op1 α
# 20t (alternate set): Transmit and hold.
#p H-9 a20.2i
#s a20.op2 A
# 18<->20
p C-3 a20.6i  # 18<->20 +0: Wait for a15 to clear.
s a20.op6 0
s a20.cc6 0
s a20.rp6 1
p a20.6o N-11
p N-11 a20.7i # 18<->20 +1: a15 += a20; a20 = 0
s a20.op7 A
s a20.cc7 C
s a20.rp7 1
p a20.7o M-1
p M-1 a20.8i  # 18<->20 +2: a20 += a18; a18 = 0
s a20.op8 β
s a20.cc8 0
s a20.rp8 1
p a20.8o M-2
# Dummy program to delay C-7 for 13l +1 (as N-8)
p C-7 a20.9i
s a20.op9 0
s a20.cc9 0
s a20.rp9 1
p a20.9o N-8
# Dummy programs for fetch sequence.
# Fig 2.2 has this on a unit annotated "SC"... which might be selective clear?
# Did they really repurpose selective clear as dummy pulses?
p F-3 a20.11i
s a20.op11 0
s a20.cc11 0
s a20.rp11 1
p a20.11o F-5
p C-5 a20.12i
s a20.op12 0
s a20.cc12 0
s a20.rp12 6
p a20.12o E-10

# Constant transmitter
# Constant transfer orders (uv) transmit 20 digits to a11 and a15, e.g. order
# AB first transmits a15 += B then sets a11 = A.
p c.o 1
p D-7 c.1i
p c.1o c.2i
s c.s1 Blr
s c.s2 Alr
p D-8 c.7i
p c.7o c.8i
s c.s7 Dlr
s c.s8 Clr
p D-9 c.13i
p c.13o c.14i
s c.s13 Flr
s c.s14 Elr
p D-10 c.19i
p c.19o c.20i
s c.s19 Hlr
s c.s20 Glr
p D-11 c.25i
p c.25o c.26i
s c.s25 Klr
s c.s26 Jlr
# Dummy program to delay C-1->J-1, i.e. xl->a15 transmit.
# (Figure 3.3 has this on the constant unit.)
p C-1 c.3i
p c.3o J-1

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

# Multiplier
# The static wiring is set up for ier=a11 and icand=a12, even though X takes
# icand in a15.
p m.ier a11
p m.icand a12
p m.L a13
p m.R a15
p m.lhppI 6
p m.rhppI 7
p m.RS M-10
p m.DS M-11
p m.F M-7
s m.ieracc1 0
s m.iercl1 0
s m.icandacc1 0
s m.icandcl1 0
s m.sf1 off
s m.place1 10
s m.prod1 A
p m.1i E-5  # X
p m.1o C-5  # retrigger fetch sequence

# Divider / square rooter
p d.quotient a4
p d.numerator a5
p d.denominator a7
p d.shift a9
p d.ans 10
s d.nr1 0
s d.nc1 0
s d.dr1 0
s d.dc1 0
s d.pl1 D10
s d.ro1 RO
s d.an1 1
s d.il1 NI
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

#p 1-2 d.1i
##p d.1o 1-3
#p 1-3 a1.3i
#p a2.A 7
#p 7 a1.γ
#s a1.op3 γ

#p d.2o 1-3
#p 1-3 a1.1i
#p 9 a1.α
#s a1.op1 α
#s a1.cc1 0

# Index of program lines
