import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:phone/core/constants/app_radius.dart';
import 'package:phone/core/widgets/card_badge.dart';
import 'package:phone/features/groups/widgets/group_card.dart';
import 'package:phone/features/groups/widgets/group_header.dart';
import 'package:phone/features/groups/widgets/group_hero_card.dart';
import 'package:phone/features/groups/widgets/group_popup_menu.dart';
import 'package:phone/generated/models/group_summary_response.dart';
import 'package:phone/i18n/strings.g.dart';

GroupSummaryResponse _group({
  String name = 'Morning Yoga',
  int? countOfStudents = 8,
  double? defaultPrice,
  int? countOfDebtors,
  String? note,
}) => GroupSummaryResponse(
  id: 'group-1',
  name: name,
  countOfStudents: countOfStudents,
  defaultPrice: defaultPrice,
  countOfDebtors: countOfDebtors,
  note: note,
  createdAt: DateTime(2026, 1, 15),
);

Future<void> _pumpHeader(
  WidgetTester tester,
  GroupSummaryResponse group,
) async {
  await tester.pumpWidget(
    TranslationProvider(
      child: MaterialApp(
        home: Scaffold(
          body: CustomScrollView(
            slivers: [
              SliverPersistentHeader(
                pinned: true,
                delegate: GroupHeaderDelegate(group: group, topPadding: 0),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: 1000)),
            ],
          ),
        ),
      ),
    ),
  );
  await tester.pump();
}

void main() {
  setUpAll(() async {
    await initializeDateFormatting();
    LocaleSettings.setLocaleSync(AppLocale.en);
  });

  group('GroupHeaderDelegate computeContentHeight', () {
    test('calculates height for basic group without note', () {
      final group = _group();
      final height = GroupHeaderDelegate.computeContentHeight(group);
      // 32 padding + max(36 avatar, 24 name + 8 spacing + 28 1-line-badges) = 32 + 60 = 92.0
      expect(height, equals(92.0));
    });

    test(
      'calculates height with defaultPrice and countOfDebtors (3 badges)',
      () {
        final group = _group(defaultPrice: 3500, countOfDebtors: 2);
        final height = GroupHeaderDelegate.computeContentHeight(group);
        // 3 badges still fit on 1 line: (3/3).ceil() = 1 -> 28.0
        expect(height, equals(92.0));
      },
    );

    test('calculates height including note and divider', () {
      final group = _group(note: 'Some group notes here');
      final height = GroupHeaderDelegate.computeContentHeight(group);
      // 92.0 + 25.0 (divider + margins) + note text height (~20.0)
      expect(height, greaterThan(130.0));
    });
  });

  group('GroupHeaderDelegate rendering', () {
    testWidgets(
      'displays GroupHeroCard with badges and actions when uncollapsed',
      (tester) async {
        final group = _group(
          name: 'Morning Yoga',
          countOfStudents: 12,
          defaultPrice: 5000,
          countOfDebtors: 3,
        );

        await _pumpHeader(tester, group);

        expect(find.byType(GroupHeroCard), findsOneWidget);
        expect(find.byType(GroupCard), findsOneWidget);
        expect(find.text('Morning Yoga'), findsOneWidget);
        expect(
          find.descendant(
            of: find.byType(CardBadge),
            matching: find.text('12'),
          ),
          findsOneWidget,
        );
        expect(
          find.descendant(of: find.byType(CardBadge), matching: find.text('3')),
          findsOneWidget,
        );
        expect(find.byType(BackButton), findsOneWidget);
        expect(find.byType(GroupPopupMenu), findsOneWidget);
      },
    );

    testWidgets('shows note when note is provided', (tester) async {
      final group = _group(note: 'Tuesdays and Thursdays');

      await _pumpHeader(tester, group);

      expect(find.text('Tuesdays and Thursdays'), findsOneWidget);
    });

    testWidgets(
      'border radius is zero when uncollapsed and bottomRounded when collapsed',
      (tester) async {
        await _pumpHeader(tester, _group());

        // Uncollapsed: border radius is zero
        final uncollapsedMaterial = tester.widget<Material>(
          find
              .descendant(
                of: find.byType(SliverPersistentHeader),
                matching: find.byType(Material),
              )
              .first,
        );
        final uncollapsedBorder =
            uncollapsedMaterial.shape as RoundedRectangleBorder;
        expect(uncollapsedBorder.borderRadius, equals(BorderRadius.zero));

        // Scroll to collapse
        await tester.drag(find.byType(CustomScrollView), const Offset(0, -300));
        await tester.pumpAndSettle();

        // Collapsed: border radius is bottomRounded
        final collapsedMaterial = tester.widget<Material>(
          find
              .descendant(
                of: find.byType(SliverPersistentHeader),
                matching: find.byType(Material),
              )
              .first,
        );
        final collapsedBorder =
            collapsedMaterial.shape as RoundedRectangleBorder;
        expect(collapsedBorder.borderRadius, equals(AppRadius.bottomRounded));

        // Pinned header shows group name in collapsed state
        expect(find.text('Morning Yoga'), findsOneWidget);
      },
    );
  });
}
