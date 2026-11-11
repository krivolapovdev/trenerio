import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:phone/features/settings/pages/settings_page.dart';
import 'package:phone/i18n/strings.g.dart';

void main() {
  setUpAll(() => LocaleSettings.setLocaleSync(AppLocale.en));

  testWidgets('settings page does not contain temporary rosters', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        child: TranslationProvider(
          child: const MaterialApp(home: SettingsPage(title: 'Settings')),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Temporary rosters'), findsNothing);
    expect(find.text('Appearance'), findsOneWidget);
    expect(find.text('Account'), findsOneWidget);
    expect(find.text('Other'), findsOneWidget);
  });
}
