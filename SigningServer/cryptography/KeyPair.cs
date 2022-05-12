using System.Numerics;
using Cryptography.EllipticCurveCryptography;

namespace Cryptography
{
    public class KeyPair : ECPoint
    {
        private BigInteger private_key;
        private Curves used_curve;
        public Curves curve { get { return used_curve; } }

        public BigInteger private_component { get { return private_key; } }
        public Coordinate public_component { get { return getCoords(); } }

        public KeyPair(Curves curve) : base(curve)
        {
            used_curve = curve;
            private_key = ECC.randomBigInteger(1, order - 1);

            multiply(private_component);
        }

        public KeyPair(Curves curve, BigInteger private_key) : base(curve)
        {
            used_curve = curve;
            this.private_key = private_key;

            multiply(private_component);
        }
    }
}
