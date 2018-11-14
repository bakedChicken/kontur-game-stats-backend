using LiteDB;

namespace Kontur.GameStats.Server.Models {
    public class DirtyServer {
        [BsonId]
        public string Endpoint { get; set; }
    }
}
