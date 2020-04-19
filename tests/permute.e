set a20 9876543210

p 1-1 a20.1i
s a20.op1 A
p a20.A 1

# reverse digits of a20 preserving sign
p	1 ad.permute.1.11,1,2,3,4,5,6,7,8,9,10
p ad.permute.1.11,1,2,3,4,5,6,7,8,9,10 a19.α
p 1-1 a19.1i
s a19.op1 α

# duplicate digits and delete some
p	1 ad.permute.2.0,1,1,2,2,3,3,4,4,5,0
p ad.permute.2.0,1,1,2,2,3,3,4,4,5,0 a18.α
p 1-1 a18.1i
s a18.op1 α

# merge several permutations using a trunk
p	1 ad.permute.3.0,0,0,0,0,0,0,0,0,2,1
p ad.permute.3.0,0,0,0,0,0,0,0,0,2,1 2
p	1 ad.permute.4.0,0,0,0,0,0,0,4,3,0,0
p ad.permute.4.0,0,0,0,0,0,0,4,3,0,0 2
p 2 a17.α
p 1-1 a17.1i
s a17.op1 α

p i.io 1-1
b i
