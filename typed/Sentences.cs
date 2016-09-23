using System.Collections.Generic;
using System.Drawing;
using System.IO;

namespace typed
{
    public class Sentences
    {
        public string SentenceNumber { get; set; }
        public FileInfo CursorFull { get; set; }
        public FileInfo CursorEmpty { get; set; }
        public IEnumerable<FileInfo> SentenceFile { get; set; }
    }
}