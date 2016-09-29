using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using AForge.Video.DirectShow;
using AForge.Video.FFMPEG;
using System.Text.RegularExpressions;

namespace typed
{
    class Program
    {
        static void Main(string[] args)
        {
            //clean strings before adding to array.
            var arrayOfSentences = new string[] { "These are the default values...", "Emilio?", "Use your own!", "Have a great day!"};

            var typly = new TyplyGenerator();

            typly.GenerateTyply(arrayOfSentences);

            //typly.MakeVideo(10);

           // typly.DeleteAllImages();

            Console.ReadLine();
        }
    }
}
