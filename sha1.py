import os
import struct
import sys 

WORDBITS = 32
def leftRotate(val, shifts):
    return ((val << shifts) | (val >> (WORDBITS - shifts))) & 0xffffffff

if __name__ == "__main__":
    if (len(sys.argv) > 1):
        if os.path.isfile(sys.argv[1]):
            message = open(sys.argv[1],'rb').read()
        else:
            message = sys.argv[1].encode('ascii')
            message += "\n"
    else:
        message = b'hi'
    mByteLength = len(message)

    h = [0X67452301, 0Xefcdab89, 0X98badcfe, 0X10325476, 0Xc3d2e1f0]
    # append bit 1 to message as a byte
    message += b'\x80'

    # append k bits so that message length is 56 modulo 64 (missing 64 bits, 8 bytes)
    message += b'\x00' * ((56 - (mByteLength+1))%64)

    # append the length of the original message (in bits) in Big-Endian order, forcing 8 bytes
    message += struct.pack(">Q",mByteLength*8)

    # break message into 512 bit (64 byte) chunks
    chunks = [message[i:i+64] for i in range(0,len(message),64)]

    for chunk in chunks:
        w = [0]*80
        # initialize hash value for this chunk:
        a = h[0]
        b = h[1]
        c = h[2]
        d = h[3]
        e = h[4]

        for i in range(80):
            if (i < 16):
                # break chunk into 16 32-bit words 
                w[i] = struct.unpack(b">I",chunk[4*i:4*i+4])[0]
            else:
                # extend the 16 words into 80 32-bit words
                w[i] = leftRotate(w[i-3] ^ w[i-8] ^ w[i-14] ^ w[i-16],1)

            if i < 20:
                f = d ^ (b & (c ^ d))
                k = 0x5a827999 
            elif i < 40:
                f = b ^ c ^ d
                k = 0x6ed9eba1
            elif i < 60:
                f = (b & c) | (b & d) | (c & d)
                k = 0x8f1bbcdc
            else:
                f = b ^ c ^ d
                k = 0xca62c1d6

            t = (leftRotate(a, 5) + f + e + k + w[i]) & 0xffffffff
            e = d
            d = c
            c = leftRotate(b,30)
            b = a
            a = t

        # add this chunk's hash to the result so far:
        h[0] = (h[0] + a) & 0xffffffff
        h[1] = (h[1] + b) & 0xffffffff
        h[2] = (h[2] + c) & 0xffffffff
        h[3] = (h[3] + d) & 0xffffffff
        h[4] = (h[4] + e) & 0xffffffff

    print("sha-1 digest: %08x%08x%08x%08x%08x" % tuple(h))
