using System.Numerics;
using System;

namespace Cryptography
{
    public static class MathBI
    {
        //Calculates the remainder of any division, including negative divisions
        public static BigInteger mod(BigInteger number, BigInteger modulus)
        {
            if (modulus == 0)
            {
                return number;
            }

            return (number % modulus + modulus) % modulus;
        }

        //Modular multiplicative inverse
        public static BigInteger mmi(BigInteger number, BigInteger modulus)
        {
            return BigInteger.ModPow(number, modulus - 2, modulus);
        }
    }
}
