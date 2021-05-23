set a20 9876543210
set a15 -9876543210

p 1-1 a20.1i
s a20.op1 A
p a20.A 1

p 1-1 a15.1i
s a15.op1 AS
p a15.A 3

# reverse digits of a20 preserving sign
s ad.permute.1 11,1,2,3,4,5,6,7,8,9,10
p	1 ad.permute.1
p ad.permute.1 a19.α
p 1-1 a19.1i
s a19.op1 α

# duplicate digits and delete some
p	1 ad.permute.2.0,1,1,2,2,3,3,4,4,5,0
p ad.permute.2.0,1,1,2,2,3,3,4,4,5,0 a18.α
p 1-1 a18.1i
s a18.op1 α

# merge several permutations using a trunk
s ad.permute.3 0,0,0,0,0,0,0,0,0,2,1
s ad.permute.4 0,0,0,0,0,0,0,4,3,0,0
p	1 ad.permute.3
p ad.permute.3 2
p	1 ad.permute.4
p ad.permute.4 2
p 2 a17.α
p 1-1 a17.1i
s a17.op1 α

# duplicate PM digit
s ad.permute.5 11,11,11,8,7,6,5,4,3,2,1
p 3 ad.permute.5
p ad.permute.5 a16.α
p 1-1 a16.1i
s a16.op1 α

p i.io 1-1
b i
