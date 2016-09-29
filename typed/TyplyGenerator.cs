using AForge.Video.FFMPEG;
using Gif.Components;
using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace typed
{
    public class TyplyGenerator
    {
        private int _width = 1280;
        private int _height = 720;
        private string _path = "D:\\Typly";

        public TyplyGenerator()
        {

        }
        public TyplyGenerator(string path)
        {
            _path = path;
        }

        public TyplyGenerator(int width, int height)
        {
            _width = width;
            _height = height;
        }

        public TyplyGenerator(int width, int height, string path)
        {
            _width = width;
            _height = height;
            _path = path;
        }

        public void GenerateTyply(string[] sentences)
        {
            for (int i = 0; i <= sentences.Length - 1; i++)
            {
                Typed(CleanText(sentences[i]), "", 0, i);
            }
        }

        /// <summary>
        /// Generate the png files character by character
        /// </summary>
        /// <param name="sentence"></param>
        /// <param name="text"></param>
        /// <param name="curStrPos"></param>
        /// <param name="currentSentence"></param>
        public void Typed(string sentence, string text, int curStrPos = 0, int currentSentence = 0)
        {
            //finished sentence
            if (curStrPos > sentence.Length - 1)
            {
                return;
            }
            // start typing each new char into existing string
            // curString: arg, self.text: original text inside element
            text = text + sentence[curStrPos];
            var cursor = "|";

            //last character so we need to add an space
            if (curStrPos == sentence.Length - 1)
            {
                cursor = " |";

                var cursorOff = DrawText(text, _width);
                cursorOff.Save(string.Format(_path + "\\{0}-{1}.png", currentSentence, curStrPos + 1), ImageFormat.Png);

                var cursorOn = DrawText(text + cursor, _width);
                cursorOn.Save(string.Format(_path + "\\{0}-{1}.png", currentSentence, curStrPos + 2), ImageFormat.Png);

                cursorOff.Save(string.Format(_path + "\\{0}-{1}.png", currentSentence, curStrPos + 3), ImageFormat.Png);

                cursorOn.Save(string.Format(_path + "\\{0}-{1}.png", currentSentence, curStrPos + 4), ImageFormat.Png);

            }

            var characterImage = DrawText(text + cursor, _width);
            characterImage.Save(string.Format(_path + "\\{0}-{1}.png", currentSentence, curStrPos), ImageFormat.Png);

            // add characters one by one
            curStrPos++;
            
            // loop the function
            Typed(sentence, text, curStrPos, currentSentence);
        }

       /// <summary>
       /// get dimensions, width and height of the sentence, based on the font type and size.
       /// </summary>
       /// <param name="sentence"></param>
       /// <returns></returns>
        private ImageDimensions GetDimensions(string sentence)
        {
            Image img = new Bitmap(1, 1);
            Graphics drawing = Graphics.FromImage(img);

            //add cursor to the end of the sentence to add the needed space
            var sentenceWithCursor = sentence + " |";
            //measure the string to see how big the image needs to be
            SizeF textSize = drawing.MeasureString(sentenceWithCursor, new Font("Times", 72));

            var width = (int)textSize.Width;
            var height = (int)textSize.Height;

            if ((width & 1) != 0)
            {
                width += 1;
            }
            if ((height & 1) != 0)
            {
                height += 1;
            }
            return new ImageDimensions { Width = width, Height = height };
        }
        /// <summary>
        /// Draw the text on the bitmap
        /// </summary>
        /// <param name="text">text to add to the image</param>
        /// <param name="xPosition">position of the text</param>
        /// <param name="colorOpacity">opacity of the text brush</param>
        /// <returns></returns>
        private Image DrawText(string text, int xPosition = 1, int colorOpacity = 255)
        {
            //first, create a dummy bitmap just to get a graphics object
            System.Drawing.Image img = new Bitmap(1, 1);
            Graphics drawing = Graphics.FromImage(img);

            //measure the string to see how big the image needs to be
            SizeF textSize = drawing.MeasureString(text, new Font("Times", 72));

            //free up the dummy image and old graphics object
            img.Dispose();
            drawing.Dispose();

            //create a new image of the right size
            img = new Bitmap(_width, _height);

            drawing = Graphics.FromImage(img);

            //paint the background
            drawing.Clear(Color.White);

            //create a brush for the text
            Color newColor = Color.FromArgb(colorOpacity, Color.Black);
            Brush textBrush = new SolidBrush(newColor);

            // Create rectangle for drawing.
            // with this word wrapping is easier
            float x = 1;
            float y = 0;
            float width = _width;
            float height = _height;

            RectangleF drawRect = new RectangleF(x, y, width, height);

            drawing.DrawString(text, new Font("Times", 72), textBrush, drawRect);

            drawing.Save();

            textBrush.Dispose();
            drawing.Dispose();

            return img;

        }

        public IEnumerable<Sentences> GetSentences()
        {
            var di = new DirectoryInfo(_path);
            var images = di.GetFiles().Where(c => c.Extension == ".png").Select(v => new { SentenceId = v.Name.Split('-')[0], Info = v });

            var grouped = images.GroupBy(b => b.SentenceId);
            var list = new List<Sentences>();

            foreach (var item in grouped)
            {
                var sentence = new Sentences();
                sentence.SentenceNumber = item.Key;
                sentence.CursorFull = item.Where(b => b.Info.Name.Contains("cursorFull")).SingleOrDefault().Info;
                sentence.CursorEmpty = item.Where(b => b.Info.Name.Contains("cursorEmpty")).SingleOrDefault().Info;
                sentence.SentenceFile = item.Where(b => !b.Info.Name.Contains("cursor")).Select(b => b.Info)
                    .OrderBy(b => int.Parse(Path.GetFileNameWithoutExtension(b.Name.Split('-')[1])));
                list.Add(sentence);
            }

            return list;
        }
        /// <summary>
        /// Generate a typly gif
        /// </summary>
        public void MakeGif()
        {
            string gifName = _path + "\\typly.gif";

            AnimatedGifEncoder e = new AnimatedGifEncoder();
            e.Start(gifName);
            e.SetFrameRate(12);
            e.SetRepeat(0); //-1:no repeat,0:always repeat 

            var images = GetSentences();

            Sentences currentSentence = null;
            foreach (var item in images)
            {
                if (currentSentence != null && item.SentenceNumber != currentSentence.SentenceNumber)
                {
                    foreach (var reversedImage in currentSentence.SentenceFile.Reverse())
                    {
                        e.AddFrame(new Bitmap(reversedImage.FullName));
                        
                    }
                }

                foreach (var image in item.SentenceFile)
                {
                    e.AddFrame(new Bitmap(image.FullName));
                }

                SetCursorBlink(e, item, 2);
                currentSentence = item;
            }
            //add last frame of cursor blinking for 2 seconds
            var twoCursors = images.Last();

            SetCursorBlink(e, twoCursors, 6);

            e.Finish();
        }
        /// <summary>
        /// Generate a .mp4 typly video 
        /// </summary>
        /// <param name="framesPerSecond"></param>
        public void MakeVideo(int framesPerSecond = 10)
        {
            VideoFileWriter writer = new VideoFileWriter();

            writer.Open(_path + "\\typly.mp4", _width, _height, framesPerSecond, VideoCodec.MPEG4, 1000000);

            var images = GetSentences().ToList();

            Sentences currentSentence = null;
            foreach (var item in images)
            {
                if (currentSentence != null && item.SentenceNumber != currentSentence.SentenceNumber)
                {
                    foreach (var reversedImage in currentSentence.SentenceFile.Reverse())
                    {
                        writer.WriteVideoFrame(LoadImageToMemory(reversedImage.FullName));
                    }

                }

                foreach (var image in item.SentenceFile)
                {
                    writer.WriteVideoFrame(LoadImageToMemory(image.FullName));
                }

                SetCursorBlink(writer, item, 2);
                currentSentence = item;
            }

           //add last frame of cursor blinking for 2 seconds
           var twoCursors = images.Last();
           SetCursorBlink(writer, twoCursors, 6);

            writer.Close();
        }
        /// <summary>
        /// generate multiple frames of the cursor (|) based on a loop iterations
        /// </summary>
        /// <param name="writer">VideoFileWriter (open) for adding the frames</param>
        /// <param name="sentence">Sentence object with the empty and full cursor png images</param>
        /// <param name="interations">number of iterations 2 iterations equal to half a second frame</param>
        private void SetCursorBlink(VideoFileWriter writer, Sentences sentence, int interations)
        {
            var emptyCursor = sentence.CursorEmpty;
            var fullCursor = sentence.CursorFull;

            for (int i = 0; i < interations; i++)
            {
                writer.WriteVideoFrame(LoadImageToMemory(emptyCursor.FullName));
                writer.WriteVideoFrame(LoadImageToMemory(emptyCursor.FullName));

                writer.WriteVideoFrame(LoadImageToMemory(fullCursor.FullName));
                writer.WriteVideoFrame(LoadImageToMemory(fullCursor.FullName));
            }
        }

        /// <summary>
        /// generate multiple frames of the cursor (|) based on a loop iterations
        /// </summary>
        /// <param name="writer">VideoFileWriter (open) for adding the frames</param>
        /// <param name="sentence">Sentence object with the empty and full cursor png images</param>
        /// <param name="interations">number of iterations 2 iterations equal to half a second frame</param>
        private void SetCursorBlink(AnimatedGifEncoder writer, Sentences sentence, int interations)
        {
            var emptyCursor = sentence.CursorEmpty;
            var fullCursor = sentence.CursorFull;

            for (int i = 0; i < interations; i++)
            {
                var emptyCursorFrame = new Bitmap(emptyCursor.FullName);
                writer.AddFrame(emptyCursorFrame);
                writer.AddFrame(emptyCursorFrame);

                var fullCursorFrame = new Bitmap(fullCursor.FullName);
                writer.AddFrame(fullCursorFrame);
                writer.AddFrame(fullCursorFrame);
            }

        }

        public void DeleteAllImages()
        {
            var di = new DirectoryInfo(_path);
            var images = di.GetFiles().Where(c => c.Extension == ".png").ToList();
            foreach (var item in images)
            {
                File.Delete(item.FullName);
            }
        }

        private string CleanText(string text)
        {
            var cleaned = text.Replace("\n", "").Replace("\t", "").Replace("\r", "");

            return cleaned;
        }

        private Bitmap LoadImageToMemory(string imagePath)
        {
            using (StreamReader streamReader = new StreamReader(imagePath))
            {
                Bitmap tmpBitmap = (Bitmap)Bitmap.FromStream(streamReader.BaseStream);
                return tmpBitmap;
            }
        }
    }
}
