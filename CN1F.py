class CN1FAddr:
    def __init__(self, SrcAddr, DstAddr, Address, MultiMsg, A8000, A10000, SeqNr):
        self.SrcAddr = SrcAddr
        self.DstAddr = DstAddr
        self.Address = Address
        self.MultiMsg = MultiMsg
        self.A8000 = A8000
        self.A10000 = A10000
        self.SeqNr = SeqNr

        self.Seq = 1

    @classmethod
    def fromCanID(cls, CID):
        return cls(
            (CID >> 0) & 0x3f,
            (CID >> 6) & 0x3f,
            (CID >> 12) & 0x03,
            (CID >> 14) & 0x01,
            (CID >> 15) & 0x01,
            (CID >> 16) & 0x01,
            (CID >> 17) & 0x03)

    def CanID(self):
        addr = 0x0
        addr |= self.SrcAddr << 0
        print(addr)
        addr |= self.DstAddr << 6
        print(addr)

        addr |= self.Address << 12
        print(addr)
        addr |= self.MultiMsg << 14
        print(addr)
        addr |= self.A8000 << 15
        print(addr)
        addr |= self.A10000 << 16
        print(addr)
        addr |= self.SeqNr << 17
        print(addr)
        addr |= 0x1F << 24
        print(addr)

        return addr

    def canwrite(self, msg, data=[]):
        print("writing", msg, data)

    def write_CN_Msg(self, data):
        print("data", data)

        print("seq before", self.Seq)
        self.Seq = (self.Seq + 1) & 0x3
        print("seq after", self.Seq)

        print("len", len(data))
        if len(data) > 8:
            datagrams = int(len(data) / 7)
            print("datagrams", len(data) / 7)
            if len(data) == datagrams * 7:
                print("datagrams decreaed by 1")
                datagrams -= 1
            for n in range(datagrams):
                print("n", n)
                self.canwrite(self.CanID(), [n] + data[n * 7:n * 7 + 7])
            n += 1
            restlen = len(data) - n * 7
            print("restlen", restlen)
            print("rest data", data[n * 7:n * 7 + restlen])
            self.canwrite(self.CanID(), [n | 0x80] + data[n * 7:n * 7 + restlen])
        else:
            self.canwrite(self.CanID(), data)


def get_can_id(can_id):
    if can_id & 0xFFFFFFC0 == 0x10000000:
        comfoAddr = can_id & 0x3f
        print(f'{comfoAddr:#06X}')
        print(comfoAddr)


cn1f = CN1FAddr(11, 1, 0x1, 0x0, 0x0, 0x1, 0x1)
print(type(cn1f.CanID()))
print("CANID")
print(cn1f.CanID())
print(hex(cn1f.CanID()))

addr = 0x1F0752CC
print(addr)
print("Src", CN1FAddr.fromCanID(addr).SrcAddr)
print("Dst", CN1FAddr.fromCanID(addr).DstAddr)
print("Address", CN1FAddr.fromCanID(addr).Address)
print("MultiMsg", CN1FAddr.fromCanID(addr).MultiMsg)
print("A8000", CN1FAddr.fromCanID(addr).A8000)
print("A10000", CN1FAddr.fromCanID(addr).A10000)
print("Seq", CN1FAddr.fromCanID(addr).SeqNr)

addr = 0x1F0552CC
print(addr)
print("Src", CN1FAddr.fromCanID(addr).SrcAddr)
print("Dst", CN1FAddr.fromCanID(addr).DstAddr)
print("Address", CN1FAddr.fromCanID(addr).Address)
print("MultiMsg", CN1FAddr.fromCanID(addr).MultiMsg)
print("A8000", CN1FAddr.fromCanID(addr).A8000)
print("A10000", CN1FAddr.fromCanID(addr).A10000)
print("Seq", CN1FAddr.fromCanID(addr).SeqNr)

# cn1f.write_CN_Msg([0x84, 0x15, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1C, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00])


get_can_id(0x10000011)
