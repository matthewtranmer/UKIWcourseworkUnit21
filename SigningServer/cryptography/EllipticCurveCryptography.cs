using System;
using System.Numerics;
using System.Text;
using System.Security.Cryptography;
using System.Globalization;

namespace Cryptography.EllipticCurveCryptography
{
    public class ECC
    {
        private Curves Curve;
        public Curves curve { get { return Curve; } }

        public ECC(Curves curve)
        {
            Curve = curve;
        }

        //Generates positive big integer between and including min and max
        //NOT FOR PRODUCTION - TESTING ONLY
        static public BigInteger randomBigInteger(BigInteger min, BigInteger max)
        {
            int length = Convert.ToInt32(BigInteger.Log10(max) + 1);
            byte[] buffer = new byte[length];

            Random random = new Random();
            random.NextBytes(buffer);

            BigInteger number = new BigInteger(buffer, true);
            number = (number + 1) % (max - min + 1) + min;

            return number;
        }

        private string generateHash(string data)
        {
            using SHA256 sha256Hash = SHA256.Create();
            byte[] data_bytes = Encoding.UTF8.GetBytes(data);
            byte[] hash_bytes = sha256Hash.ComputeHash(data_bytes);

            return BitConverter.ToString(hash_bytes).Replace("-", String.Empty);
        }

        private string KDF(string data, int key_length)
        {
            return generateHash(data).Substring(0, key_length);
        }

        //Generate a shared secret with a private component and a public elliptic curve point
        public string ECDH(BigInteger private_component, Coordinate public_component, int key_length = 32)
        {
            ECPoint point = new ECPoint(Curve);
            point.changePoint(public_component);

            point.multiply(private_component);

            return KDF(Convert.ToString(point.getCoords().x), key_length);
        }

        //OPTIMIZE (div operator)
        private string uintToBinary(BigInteger full_hash_int)
        {
            string hash_binary = "";
            while (full_hash_int != 0)
            {
                if (full_hash_int % 2 == 1)
                {
                    hash_binary = "1" + hash_binary;
                }
                else
                {
                    hash_binary = "0" + hash_binary;
                }

                full_hash_int /= 2;
            }

            return hash_binary;
        }

        private BigInteger ubinaryToInt(string binary)
        {
            BigInteger decimal_value = 0;

            int count = 0;
            for (int i = binary.Length - 1; i >= 0; i--)
            {
                if (binary[i] == '1')
                {
                    decimal_value += BigInteger.Pow(2, count);
                }

                count++;
            }

            return decimal_value;
        }

        private BigInteger convertHash(string hash, BigInteger order)
        {
            BigInteger int_hash = BigInteger.Parse("0" + hash, NumberStyles.AllowHexSpecifier);
            string hash_binary = uintToBinary(int_hash);
            int binary_hash_length = Convert.ToInt32(Math.Floor(BigInteger.Log(order, 2) + 1));

            string shortened_binary = hash_binary.Substring(0, binary_hash_length);
            BigInteger shortened_hash = ubinaryToInt(shortened_binary);

            return shortened_hash;
        }

        private string coordinateToString(Coordinate coord){
            return $"{coord.x.ToString("x")},{coord.y.ToString("x")}";
        }

        private Coordinate stringToCoordinate(string str_coord){
            string[] split_string = str_coord.Split(',', 2);

            Coordinate coordinate = new Coordinate(
                BigInteger.Parse(split_string[0], NumberStyles.AllowHexSpecifier),
                BigInteger.Parse(split_string[1], NumberStyles.AllowHexSpecifier)
            );
            return coordinate;
        }

        public (string signature, string public_key) generateDSAsignature(string data, BigInteger private_key)
        {
            KeyPair key_pair = new KeyPair(curve, private_key);

            BigInteger order = key_pair.getOrder();
            string hash = generateHash(data);
            BigInteger int_hash = convertHash(hash, order);

            BigInteger ephemeral_key = 0;
            BigInteger r = 0;
            BigInteger s = 0;

            while (s == 0)
            {
                while (r == 0)
                {
                    ephemeral_key = randomBigInteger(1, order - 1);
                    ECPoint ephemeral_point = new ECPoint(curve);
                    ephemeral_point.multiply(ephemeral_key);

                    r = MathBI.mod(ephemeral_point.getCoords().x, order);
                }

                s = int_hash + (r * key_pair.private_component);
                s = s * MathBI.mmi(ephemeral_key, order);
                s = MathBI.mod(s, order);
            }

            string signature = $"{r.ToString("x")}:{s.ToString("x")}";
            string public_key = coordinateToString(key_pair.public_component);
            return (signature, public_key);
        }
        public bool verifyDSAsignature(string data, string signature, string public_key_str)
        {
            Coordinate public_key = stringToCoordinate(public_key_str);

            ECPoint public_point = new ECPoint(curve);
            public_point.changePoint(public_key);

            if (public_point.identity_element.Equals(public_key))
            {
                return false;
            }

            if (!public_point.validatePoint())
            {
                return false;
            }

            BigInteger order = ECPoint.pre_defined_curves[curve].order;
            string hash = generateHash(data);
            BigInteger int_hash = convertHash(hash, order);

            string[] split_signature = signature.Split(':');
            BigInteger r = BigInteger.Parse(split_signature[0], NumberStyles.AllowHexSpecifier);
            BigInteger s = BigInteger.Parse(split_signature[1], NumberStyles.AllowHexSpecifier);

            if (r > order || r < 1)
            {
                return false;
            }

            if (s > order || s < 1)
            {
                return false;
            }

            BigInteger inverse_s = MathBI.mmi(s, order);

            BigInteger u1 = MathBI.mod(int_hash * inverse_s, order);
            BigInteger u2 = MathBI.mod(r * inverse_s, order);

            ECPoint point1 = new ECPoint(Curve);
            point1.multiply(u1);

            ECPoint point2 = new ECPoint(Curve);
            point2.changePoint(public_key);
            point2.multiply(u2);

            point1.addPoint(point2);

            if (point1.identity_element.Equals(point1))
            {
                return false;
            }

            BigInteger x = MathBI.mod(point1.getCoords().x, order);
            r = MathBI.mod(r, order);

            if (r == x)
            {
                return true;
            }

            return false;
        }
    }
}
