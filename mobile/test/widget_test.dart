import 'package:flutter_test/flutter_test.dart';

import 'package:typly_mobile/main.dart';

void main() {
  testWidgets('editor renders the default text and export actions', (
    tester,
  ) async {
    await tester.pumpWidget(const TyplyApp());

    expect(find.text('Create a typing animation'), findsOneWidget);
    expect(find.text('Export GIF'), findsOneWidget);
    expect(find.text('Export MP4'), findsOneWidget);
    expect(find.textContaining('Welcome to Typly'), findsOneWidget);
  });
}
