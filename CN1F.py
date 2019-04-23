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
        if (CID >> 24) != 0x1F:
            raise ValueError('Not 0x1F CMD!')
        else:
            return cls(
                (CID >> 0) & 0x3f,
                (CID >> 6) & 0x3f,
                (CID >> 12) & 0x03,
                (CID >> 14) & 0x01,
                (CID >> 15) & 0x01,
                (CID >> 16) & 0x01,
                (CID >> 17) & 0x03)

    def __repr__(self):
        return (f'{self.__class__.__name__}(\n'
                f'  SrcAddr = {self.SrcAddr:#02x},\n'
                f'  DstAddr = {self.DstAddr:#02x},\n'
                f'  Address = {self.Address:#02x},\n'
                f'  MultiMsg = {self.MultiMsg:#02x},\n'
                f'  A8000 = {self.A8000:#02x},\n'
                f'  A10000 = {self.A10000:#02x},\n'
                f'  SeqNr = {self.SeqNr:#02x})')

    def CanID(self):
        addr = 0x0
        addr |= self.SrcAddr << 0
        addr |= self.DstAddr << 6

        addr |= self.Address << 12
        addr |= self.MultiMsg << 14
        addr |= self.A8000 << 15
        addr |= self.A10000 << 16
        addr |= self.SeqNr << 17
        addr |= 0x1F << 24

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


cn1f = CN1FAddr(42, 1, 1, 1, 0, 1, 3)
print(type(cn1f.CanID()))
print(cn1f.CanID())
print(hex(cn1f.CanID()))

cn1f.write_CN_Msg([0x84, 0x15, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1C, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00])
