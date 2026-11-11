import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:phone/features/rosters/pages/roster_list_page.dart';
import 'package:phone/features/students/pages/student_payments_page.dart';
import 'package:phone/features/students/widgets/create_student_bottom_sheet.dart';
import 'package:phone/features/students/widgets/student_list_popup_menu.dart';
import 'package:phone/i18n/strings.g.dart';

void main() {
  setUpAll(() => LocaleSettings.setLocaleSync(AppLocale.en));

  Widget buildMenu() => ProviderScope(
    child: TranslationProvider(
      child: const MaterialApp(
        home: Scaffold(
          appBar: PreferredSize(
            preferredSize: Size.fromHeight(56),
            child: StudentListPopupMenu(),
          ),
        ),
      ),
    ),
  );

  testWidgets('renders popup menu with options', (tester) async {
    await tester.pumpWidget(buildMenu());
    await tester.pumpAndSettle();

    final menuButton = find.byType(PopupMenuButton<String>);
    expect(menuButton, findsOneWidget);

    await tester.tap(menuButton);
    await tester.pumpAndSettle();

    expect(find.text('Add Student'), findsOneWidget);
    expect(find.text('Temporary rosters'), findsOneWidget);
    expect(find.text('Payments'), findsOneWidget);
    expect(find.byIcon(FluentIcons.person_add_24_regular), findsOneWidget);
    expect(find.byIcon(Icons.format_list_bulleted_outlined), findsOneWidget);
    expect(find.byIcon(FluentIcons.money_24_regular), findsOneWidget);
  });

  testWidgets('tapping Add Student opens CreateStudentBottomSheet', (
    tester,
  ) async {
    await tester.pumpWidget(buildMenu());
    await tester.pumpAndSettle();

    await tester.tap(find.byType(PopupMenuButton<String>));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Add Student'));
    await tester.pumpAndSettle();

    expect(find.byType(CreateStudentBottomSheet), findsOneWidget);
  });

  testWidgets('tapping Temporary rosters opens RosterListPage', (tester) async {
    await tester.pumpWidget(buildMenu());
    await tester.pumpAndSettle();

    await tester.tap(find.byType(PopupMenuButton<String>));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Temporary rosters'));
    await tester.pumpAndSettle();

    expect(find.byType(RosterListPage), findsOneWidget);
  });

  testWidgets('tapping Payments opens StudentPaymentsPage', (tester) async {
    await tester.pumpWidget(buildMenu());
    await tester.pumpAndSettle();

    await tester.tap(find.byType(PopupMenuButton<String>));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Payments'));
    await tester.pumpAndSettle();

    expect(find.byType(StudentPaymentsPage), findsOneWidget);
  });
}
