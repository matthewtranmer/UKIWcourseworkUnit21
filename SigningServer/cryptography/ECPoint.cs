using System.Collections.Generic;
using System.Numerics;
using System.Globalization;

namespace Cryptography
{
    public struct Curve
    {
        public Coordinate generator;
        public BigInteger a;
        public BigInteger b;
        public BigInteger modulus;
        public BigInteger order;

        public Curve(Coordinate generator, BigInteger a, BigInteger b, BigInteger modulus, BigInteger order)
        {
            this.generator = generator;
            this.a = a;
            this.b = b;
            this.modulus = modulus;
            this.order = order;
        }
    }

    public struct Coordinate
    {
        public BigInteger x;
        public BigInteger y;

        public Coordinate(BigInteger x, BigInteger y)
        {
            this.x = x;
            this.y = y;
        }
    }

    public enum Curves
    {
        microsoft_160
    }

    //Represents a point on an elliptic curve
    public class ECPoint
    {
        protected Coordinate point;
        protected BigInteger a, b, modulus, order;
        protected Coordinate generator;

        public Coordinate identity_element
        {
            get
            {
                ECPoint id_element = new ECPoint(generator, a, b, modulus, order);
                id_element.multiply(order);

                return id_element.getCoords();
            }
        }

        static public Dictionary<Curves, Curve> pre_defined_curves = new Dictionary<Curves, Curve>()
        {
            { Curves.microsoft_160, new Curve(
                new Coordinate(
                        BigInteger.Parse("08723947fd6a3a1e53510c07dba38daf0109fa120", NumberStyles.AllowHexSpecifier), //x
                        BigInteger.Parse("0445744911075522d8c3c5856d4ed7acda379936f", NumberStyles.AllowHexSpecifier) //y
                    ),
                    BigInteger.Parse("037a5abccd277bce87632ff3d4780c009ebe41497", NumberStyles.AllowHexSpecifier), //a
                    BigInteger.Parse("00dd8dabf725e2f3228e85f1ad78fdedf9328239e", NumberStyles.AllowHexSpecifier), //b
                    BigInteger.Parse("089abcdef012345672718281831415926141424f7", NumberStyles.AllowHexSpecifier), //modulus
                    BigInteger.Parse("089abcdef012345672716b26eec14904428c2a675", NumberStyles.AllowHexSpecifier) //order
                ) }
        };

        public BigInteger getModulus()
        {
            return modulus;
        }
        public BigInteger getOrder()
        {
            return order;
        }
        //Returns the current x and y coordinates
        public Coordinate getCoords()
        {
            return point;
        }

        //Creates a curve with pre defined parameters
        public ECPoint(Curves curve)
        {
            switch (curve)
            {
                case Curves.microsoft_160:
                    Curve curve_data = pre_defined_curves[Curves.microsoft_160];

                    setValues(curve_data.generator, curve_data.a, curve_data.b, curve_data.modulus, curve_data.order);
                    break;
            }
        }

        private void setValues(Coordinate generator, BigInteger a, BigInteger b, BigInteger modulus, BigInteger order)
        {
            this.generator = generator;
            point = generator;
            this.a = a;
            this.b = b;
            this.modulus = modulus;
            this.order = order;
        }

        public bool validatePoint()
        {
            //y^2 = x^3 + ax + b
            BigInteger y_squared = MathBI.mod(BigInteger.Pow(point.x, 3) + (a * point.x) + b, modulus);
            BigInteger real_y_squared = MathBI.mod(BigInteger.Pow(point.y, 2), modulus);

            if (y_squared != real_y_squared)
            {
                return false;
            }

            return true;
        }

        //Set the current point to another point
        public ECPoint(ECPoint point)
        {
            changePoint(point);

            a = point.a;
            b = point.b;
            modulus = point.modulus;
            order = point.order;
        }

        //Creates a point with null coords with curve constant a and modulus 
        public ECPoint(BigInteger a, BigInteger b, BigInteger modulus, BigInteger order)
        {
            this.a = a;
            this.b = b;
            this.modulus = modulus;
            this.order = order;
        }

        //Creates a new point with generator coords a curve constant 'a' and a modulus 
        public ECPoint(Coordinate coordinates, BigInteger a, BigInteger b, BigInteger modulus, BigInteger order)
        {
            point = coordinates;
            this.a = a;
            this.b = b;
            this.modulus = modulus;
            this.order = order;
        }

        //Set the current coords to another point's coords
        public void changePoint(ECPoint new_point)
        {
            point = new_point.getCoords();
        }

        //Set the current coords to another point's coords
        public void changePoint(Coordinate coords)
        {
            point = coords;
        }

        //Returns the gradient of the tangent on the point given
        //as an argument. Used in point doubling.
        private BigInteger calculateGradientOfTangent()
        {
            BigInteger gradient = (3 * BigInteger.Pow(point.x, 2)) + a;
            //gradient = MathBI.mod(gradient, modulus);
            gradient *= MathBI.mmi(point.y * 2, modulus);
            gradient = MathBI.mod(gradient, modulus);

            return gradient;
        }

        //Returns the gradient of the intersect of the two points
        //given as arguments. Used in point addition.
        private BigInteger calculateGradientOfIntersect(Coordinate coordinates)
        {
            BigInteger gradient = point.y - coordinates.y;
            gradient *= MathBI.mmi(point.x - coordinates.x, modulus);
            gradient = MathBI.mod(gradient, modulus);

            return gradient;
        }

        //calculates the coordinates of the inverse of the intersect of the line on the curve 
        //private (BigInteger x, BigInteger y) calculateNewCoords(BigInteger first_x, BigInteger second_x, BigInteger first_y, BigInteger gradient)
        private Coordinate calculateNewCoords(Coordinate first_coords, Coordinate second_coords, BigInteger gradient)
        {
            Coordinate newCoords = new Coordinate();

            newCoords.x = BigInteger.Pow(gradient, 2) - first_coords.x - second_coords.x;
            newCoords.x = MathBI.mod(newCoords.x, modulus);

            newCoords.y = gradient * (first_coords.x - newCoords.x) - first_coords.y;
            newCoords.y = MathBI.mod(newCoords.y, modulus);

            return newCoords;
        }

        //Doubles the point on the curve
        public void doublePoint()
        {
            BigInteger gradient = calculateGradientOfTangent();
            point = calculateNewCoords(point, point, gradient);
        }

        //Add two points on the curve.
        //public void addPoint(BigInteger x, BigInteger y)
        public void addPoint(Coordinate point)
        {
            //If both x coordinates are the same then double the point
            if (point.x == this.point.x)
            {
                doublePoint();
            }
            //Add the two points
            else
            {
                BigInteger gradient = calculateGradientOfIntersect(point);
                this.point = calculateNewCoords(this.point, point, gradient);
            }
        }

        //Adds two points on the curve.
        public void addPoint(ECPoint point)
        {
            addPoint(point.point);
        }

        //Multiples the current point by the value: multiplier
        public void multiply(BigInteger multiplier)
        {
            if (multiplier == 1)
            {
                return;
            }

            //point that will be doubled to create the exponential series
            ECPoint series_point = new ECPoint(this);
            //empty point - will be the new multipled value
            ECPoint resulting_point = new ECPoint(a, b, modulus, order);

            bool bIs_first_point = true;
            BigInteger binary_series = multiplier;


            while (binary_series != 0)
            {
                //if point needs to be added to make up the multiplier
                if (binary_series % 2 == 1)
                {
                    //if it is the first value, then set the resulting point coords to the coords of the first point
                    if (bIs_first_point)
                    {
                        resulting_point.changePoint(series_point);
                        bIs_first_point = false;
                    }
                    //if its not the first point then add the new point to the current point
                    else
                    {
                        resulting_point.addPoint(series_point);
                    }
                }

                //double series point
                series_point.doublePoint();
                binary_series /= 2;
            }

            //set point to the result of the multiplied point
            changePoint(resulting_point);
        }
    }
}
