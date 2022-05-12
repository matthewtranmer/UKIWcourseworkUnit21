using System;
using System.Net.Sockets;
using System.Net;
using System.Text;
using System.Numerics;
using Cryptography.EllipticCurveCryptography;
using System.Text.Json;
using System.IO;
using System.Collections.Generic;
using System.Threading.Channels;

namespace SigningServer{

    class Server{
        ECC ecc = new ECC(Cryptography.Curves.microsoft_160);
        //Testing private key
        BigInteger private_key = BigInteger.Parse("2344355445323434565634523454678");
        
        Dictionary<string, bool> blacklisted_tokens = new Dictionary<string, bool>();
        Channel<Dictionary<string, string>> blacklist_channel = Channel.CreateUnbounded<Dictionary<string, string>>();
        const string token_blacklist_path = "/home/matthew/Websites/UKIWcoursework/SigningServer/blacklisted_tokens.txt";

        void send(Socket socket, byte[] data){
            using BinaryWriter writer = new BinaryWriter(new NetworkStream(socket));
            writer.Write(Convert.ToUInt32(data.Length));
            writer.Write(data);
        }

        byte[] recv(Socket socket){
            using BinaryReader reader = new BinaryReader(new NetworkStream(socket));
            UInt32 length = reader.ReadUInt32();
            
            byte[] buffer = new byte[length];
            socket.Receive(buffer);

            return buffer;
        }

        Dictionary<string, string> generate(Socket conn, Dictionary<string, string> request){
            (string signature, string public_key) = ecc.generateDSAsignature(request["payload"], private_key);

            Dictionary<string, string> response = new Dictionary<string, string>(){
                {"signature", signature},
                {"public key", public_key}
            };
        
            return response;
        }

        Dictionary<string, string> verify(Socket conn, Dictionary<string, string> request){
            bool isValid = false;
            
            if(!isBlacklisted(request["payload"], request["signature"], request["public key"])){
                Console.WriteLine("Token not blacklisted");
                isValid = ecc.verifyDSAsignature(request["payload"], request["signature"], request["public key"]);
            }

            Dictionary<string, string> response = new Dictionary<string, string>(){
                {"is valid", Convert.ToString(isValid)}
            };

            return response;
        }

        bool isBlacklisted(string payload, string signature, string public_key){
            return blacklisted_tokens.ContainsKey(payload+signature+public_key);
        }

        Dictionary<string, string> blacklist(Socket conn, Dictionary<string, string> request){
            Dictionary<string, string> response;
            
            if (ecc.verifyDSAsignature(request["payload"], request["signature"], request["public key"])){
                blacklist_channel.Writer.WriteAsync(request);
                blacklisted_tokens.Add(request["payload"]+request["signature"]+request["public key"], true);
            
                response = new Dictionary<string, string>(){
                    {"success", "True"}
                };

                return response;
            }

            response = new Dictionary<string, string>(){
                {"success", "False"}
            };
            return response;
        }

        void HandleRequest(Socket conn){
            byte[] data = recv(conn);
            var request = JsonSerializer.Deserialize<Dictionary<string, string>>(data);

            Dictionary<string, string> response = new Dictionary<string, string>();
            switch (request?["command"]){
                case "generate":
                    Console.WriteLine("Generate");
                    response = generate(conn, request);
                    break;

                case "verify":
                    Console.WriteLine("Verify");
                    response = verify(conn, request);
                    break;

                case "blacklist":
                    Console.WriteLine("Blacklist");
                    response = blacklist(conn, request);
                    break;

                default:
                    throw new Exception("Command was not in the avaliable list of commands");
            }

            string json_response = JsonSerializer.Serialize(response);
            send(conn, UTF8Encoding.UTF8.GetBytes(json_response));
        }

//load blacklist from disk
        void init_blacklist(){
            using (StreamReader file = new StreamReader(token_blacklist_path)){
                while (true){
                    string? line = file.ReadLine();
                    
                    if(line == null){
                        return;
                    }

                    var token = JsonSerializer.Deserialize<Dictionary<string, string>>(line);
                    string hash = token["payload"]+token["signature"]+token["public key"];
                    blacklisted_tokens.Add(hash, true);
                }
            }
        }

        async Task blacklistIOwriter(){
            while(true){
                //create copy of dictionary so we can remove the command key 
                Dictionary<string, string> data = new Dictionary<string,string>(await blacklist_channel.Reader.ReadAsync());
                data.Remove("command");

                string serialised_data = JsonSerializer.Serialize(data);

                using (StreamWriter file = new StreamWriter(token_blacklist_path, append: true)){
                    file.WriteLine(serialised_data);
                }
            }
        }

        public void start(){
            Console.WriteLine("Starting Blacklist Writer Thread");
            Task.Factory.StartNew(blacklistIOwriter);

            Console.WriteLine("Initiating Blacklist");
            init_blacklist();
            Console.WriteLine("Blacklist Initated");
            
            IPAddress ip = new IPAddress(new byte[] {127, 0, 0, 1});
            IPEndPoint ep = new IPEndPoint(ip, 50508);
            Socket socket = new Socket(AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.IP);

            socket.Bind(ep);
            socket.Listen();

            while(true){
                Socket connection = socket.Accept();
                Console.WriteLine("Connected");
                
                //connection.Send(UTF8Encoding.UTF8.GetBytes("Hello"));
                HandleRequest(connection);
                Console.WriteLine();
            }
        }
    }

    class Program{
        static void Main(){
            Server server = new Server();
            server.start();
        }
    }
}




/*
            ECC ecc = new ECC(Cryptography.Curves.microsoft_160);
            
            //BigInteger private_key = BigInteger.Parse("823492231897324980453908745308979");
            //(string signature, string public_key) = ecc.generateDSAsignature("Data to be signed!", private_key);

            string signature = "686881636271908237438749228641124747825647434600:229074815512682766547404218447545904639088299631";
            string public_key = "626097809585934779976898897734474604382255393430,522475154971523369283646639814676849854055621586";

            Console.WriteLine(signature);
            Console.WriteLine(public_key);

            Console.WriteLine(ecc.verifyDSAsignature("Hello", signature, public_key));
            */