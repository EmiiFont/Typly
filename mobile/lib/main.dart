import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

void main() => runApp(const TyplyApp());

class TyplyApp extends StatelessWidget {
  const TyplyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Typly',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xff65d1a5)),
        useMaterial3: true,
      ),
      home: const EditorPage(),
    );
  }
}

class EditorPage extends StatefulWidget {
  const EditorPage({super.key});

  @override
  State<EditorPage> createState() => _EditorPageState();
}

class _EditorPageState extends State<EditorPage> {
  static const serverURL = String.fromEnvironment(
    'TYPLY_SERVER_URL',
    defaultValue: 'http://10.0.2.2:8080',
  );

  final _textController = TextEditingController(
    text: 'Welcome to Typly 🌍\nCreate something memorable 🚀',
  );
  double _fontSize = 56;
  int _fps = 12;
  bool _busy = false;

  @override
  void dispose() {
    _textController.dispose();
    super.dispose();
  }

  Future<void> _export(String format) async {
    setState(() => _busy = true);
    final spec = {
      'sentences': _textController.text
          .split('\n')
          .where((line) => line.trim().isNotEmpty)
          .toList(),
      'width': 1280,
      'height': 720,
      'fontSize': _fontSize,
      'fps': _fps,
      'foreground': '#000000',
      'background': '#FFFFFF',
      'cursor': '|',
      'blinks': 3,
      'emoji': 'color',
    };
    try {
      final response = await http.post(
        Uri.parse('$serverURL/api/render/$format'),
        headers: {'content-type': 'application/json'},
        body: jsonEncode(spec),
      );
      if (!mounted) return;
      final message = response.statusCode == 200
          ? '$format export ready (${response.bodyBytes.length} bytes)'
          : 'Export failed: ${response.body}';
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(message)));
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not reach Typly server: $error')),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Typly'), centerTitle: false),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            Text(
              'Create a typing animation',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text('Color emoji, smooth pacing, ready for sharing.'),
            const SizedBox(height: 20),
            TextField(
              controller: _textController,
              minLines: 8,
              maxLines: 12,
              decoration: const InputDecoration(
                labelText: 'Your text',
                hintText: 'Use one sentence per line',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 20),
            Text('Font size: ${_fontSize.round()}'),
            Slider(
              value: _fontSize,
              min: 24,
              max: 120,
              divisions: 16,
              label: _fontSize.round().toString(),
              onChanged: (value) => setState(() => _fontSize = value),
            ),
            Text('Typing speed: $_fps fps'),
            Slider(
              value: _fps.toDouble(),
              min: 6,
              max: 30,
              divisions: 8,
              label: _fps.toString(),
              onChanged: (value) => setState(() => _fps = value.round()),
            ),
            const SizedBox(height: 12),
            FilledButton.icon(
              onPressed: _busy ? null : () => _export('gif'),
              icon: const Icon(Icons.gif_box_outlined),
              label: const Text('Export GIF'),
            ),
            OutlinedButton.icon(
              onPressed: _busy ? null : () => _export('mp4'),
              icon: const Icon(Icons.movie_outlined),
              label: const Text('Export MP4'),
            ),
            if (_busy)
              const Padding(
                padding: EdgeInsets.only(top: 16),
                child: LinearProgressIndicator(),
              ),
          ],
        ),
      ),
    );
  }
}
