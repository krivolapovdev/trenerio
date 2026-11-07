import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:phone/core/widgets/card_badge.dart';
import 'package:phone/features/groups/widgets/group_card.dart';
import 'package:phone/generated/models/group_summary_response.dart';

final _group = GroupSummaryResponse(
  id: 'group-a',
  name: 'Group A',
  createdAt: DateTime(2025, 8, 22),
  note: 'Bring the notebook',
  countOfStudents: 12,
  countOfDebtors: 0,
);

Widget _wrap(GroupSummaryResponse group, {bool noteVisible = true}) =>
    MaterialApp(
      home: Scaffold(
        body: GroupCard(group: group, onTap: () {}, noteVisible: noteVisible),
      ),
    );

void main() {
  testWidgets('the card draws the note of the group', (tester) async {
    await tester.pumpWidget(_wrap(_group));

    expect(find.text('Group A'), findsOneWidget);
    expect(find.text('Bring the notebook'), findsOneWidget);
  });

  testWidgets('the note of the group can be dropped', (tester) async {
    await tester.pumpWidget(_wrap(_group, noteVisible: false));

    expect(find.text('Group A'), findsOneWidget);
    expect(find.text('Bring the notebook'), findsNothing);
  });

  testWidgets('debtor badge is NOT shown when countOfDebtors is 0 or null', (
    tester,
  ) async {
    await tester.pumpWidget(_wrap(_group));

    expect(find.byIcon(CupertinoIcons.exclamationmark_triangle), findsNothing);

    final groupWithNullDebtors = GroupSummaryResponse(
      id: 'group-b',
      name: 'Group B',
      createdAt: DateTime.now(),
      countOfStudents: 5,
      countOfDebtors: null,
    );
    await tester.pumpWidget(_wrap(groupWithNullDebtors));

    expect(find.byIcon(CupertinoIcons.exclamationmark_triangle), findsNothing);
  });

  testWidgets(
    'debtor badge is shown with icon and number when countOfDebtors > 0',
    (tester) async {
      final groupWithDebtors = GroupSummaryResponse(
        id: 'group-c',
        name: 'Group C',
        createdAt: DateTime(2025, 8, 22),
        countOfStudents: 10,
        countOfDebtors: 3,
        defaultPrice: 5000,
      );

      await tester.pumpWidget(_wrap(groupWithDebtors));

      // Badge icon should be visible
      expect(find.byIcon(CupertinoIcons.eye_fill), findsOneWidget);

      // Number of debtors should be visible
      expect(find.text('3'), findsOneWidget);

      // Number of students and price should also be present
      expect(find.text('10'), findsOneWidget);

      // Verify it's inside a CardBadge
      final badges = tester
          .widgetList<CardBadge>(find.byType(CardBadge))
          .toList();
      expect(badges.any((b) => b.label == '3'), isTrue);
    },
  );
}
