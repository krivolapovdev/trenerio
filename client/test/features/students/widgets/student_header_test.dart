import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:phone/features/students/extensions/student_balance_extension.dart';
import 'package:phone/features/students/extensions/student_status_extension.dart';
import 'package:phone/features/students/widgets/student_header.dart';
import 'package:phone/generated/models/student_payment_details_response.dart';
import 'package:phone/generated/models/student_status.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/transaction_type.dart';
import 'package:phone/i18n/strings.g.dart';

StudentSummaryResponse _student({
  DateTime? paidUntil,
  int sessionsLeft = 0,
  bool free = false,
  List<StudentStatus> statuses = const [],
}) => StudentSummaryResponse(
  id: 'student-1',
  fullName: 'Pifff',
  createdAt: DateTime(2025, 8, 22),
  free: free,
  statuses: statuses,
  paidUntil: paidUntil,
  sessionsLeft: sessionsLeft,
);

TransactionResponse _payment({DateTime? paidUntil, int? sessionsAdded}) =>
    TransactionResponse(
      id: 'tx-${paidUntil ?? sessionsAdded}',
      amount: 1000,
      date: DateTime(2026, 11, 1),
      type: TransactionType.income,
      createdAt: DateTime(2026, 11, 1),
      paymentDetails: StudentPaymentDetailsResponse(
        detailsType: 'STUDENT',
        studentId: 'student-1',
        paidUntil: paidUntil,
        sessionsAdded: sessionsAdded,
      ),
    );

Future<void> _pumpHeader(
  WidgetTester tester,
  StudentSummaryResponse student, {
  List<TransactionResponse> payments = const [],
  DateTime? now,
}) async {
  await tester.pumpWidget(
    TranslationProvider(
      child: MaterialApp(
        home: Scaffold(
          body: CustomScrollView(
            slivers: [
              SliverPersistentHeader(
                pinned: true,
                delegate: StudentHeaderDelegate(
                  student: student,
                  payments: payments,
                  topPadding: 0,
                  now: now,
                ),
              ),
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

  group('StudentSummaryHeaderBadgesExtension (first line)', () {
    testWidgets('orders present before paid', (tester) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(
                statuses: const [StudentStatus.present, StudentStatus.paid],
              );
              final badges = student.toHeaderStatusBadges(context);
              expect(badges.map((b) => b.id).toList(), ['present', 'paid']);
              return const SizedBox();
            },
          ),
        ),
      );
    });

    testWidgets('orders present before paid even if passed in reverse', (
      tester,
    ) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(
                statuses: const [StudentStatus.paid, StudentStatus.present],
              );
              final badges = student.toHeaderStatusBadges(context);
              expect(badges.map((b) => b.id).toList(), ['present', 'paid']);
              return const SizedBox();
            },
          ),
        ),
      );
    });

    testWidgets('orders missing before unpaid', (tester) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(
                statuses: const [StudentStatus.missing, StudentStatus.unpaid],
              );
              final badges = student.toHeaderStatusBadges(context);
              expect(badges.map((b) => b.id).toList(), ['missing', 'unpaid']);
              return const SizedBox();
            },
          ),
        ),
      );
    });

    testWidgets('orders present before free', (tester) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(
                statuses: const [StudentStatus.present, StudentStatus.free],
              );
              final badges = student.toHeaderStatusBadges(context);
              expect(badges.map((b) => b.id).toList(), ['present', 'free']);
              return const SizedBox();
            },
          ),
        ),
      );
    });

    testWidgets('free flag shows free payment badge', (tester) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(
                free: true,
                statuses: const [StudentStatus.present],
              );
              final badges = student.toHeaderStatusBadges(context);
              expect(badges.map((b) => b.id).toList(), ['present', 'free']);
              return const SizedBox();
            },
          ),
        ),
      );
    });

    testWidgets('shows inactive badge for inactive student', (tester) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(
                statuses: const [StudentStatus.inactive],
              );
              final badges = student.toHeaderStatusBadges(context);
              expect(badges.map((b) => b.id).toList(), ['inactive']);
              return const SizedBox();
            },
          ),
        ),
      );
    });

    testWidgets('returns empty when no statuses', (tester) async {
      await tester.pumpWidget(
        TranslationProvider(
          child: Builder(
            builder: (context) {
              final student = _student(statuses: const []);
              final badges = student.toHeaderStatusBadges(context);
              expect(badges, isEmpty);
              return const SizedBox();
            },
          ),
        ),
      );
    });
  });

  group('StudentBalanceVisibility (second line)', () {
    final fixedNow = DateTime(2026, 11, 8, 12, 0);

    test('sessions left > 0 -> always show', () {
      final visibility = StudentBalanceVisibility.of(
        _student(sessionsLeft: 5),
        const [],
        fixedNow,
      );

      expect(visibility.showSessions, isTrue);
    });

    test('sessions left <= 0 -> hide it', () {
      final zero = StudentBalanceVisibility.of(_student(sessionsLeft: 0), [
        _payment(sessionsAdded: 10),
      ], fixedNow);
      final negative = StudentBalanceVisibility.of(
        _student(sessionsLeft: -1),
        const [],
        fixedNow,
      );

      expect(zero.showSessions, isFalse);
      expect(negative.showSessions, isFalse);
    });

    test('current date < paid until -> always show', () {
      final futurePaidUntil = DateTime(2026, 11, 20);
      final visibility = StudentBalanceVisibility.of(
        _student(paidUntil: futurePaidUntil),
        const [],
        fixedNow,
      );

      expect(visibility.showPaidUntil, isTrue);
    });

    test('current date == paid until -> show', () {
      final sameDayPaidUntil = DateTime(2026, 11, 8, 0, 0);
      final visibility = StudentBalanceVisibility.of(
        _student(paidUntil: sameDayPaidUntil),
        const [],
        fixedNow,
      );

      expect(visibility.showPaidUntil, isTrue);
    });

    test('current date > paid until -> hide it', () {
      final pastPaidUntil = DateTime(2026, 11, 7);
      final visibility = StudentBalanceVisibility.of(
        _student(paidUntil: pastPaidUntil),
        [_payment(paidUntil: pastPaidUntil)],
        fixedNow,
      );

      expect(visibility.showPaidUntil, isFalse);
    });

    test('paid until null -> hide it', () {
      final visibility = StudentBalanceVisibility.of(
        _student(paidUntil: null),
        const [],
        fixedNow,
      );

      expect(visibility.showPaidUntil, isFalse);
    });

    test('free student hides both balance badges', () {
      final free = StudentBalanceVisibility.of(
        _student(paidUntil: DateTime(2026, 12, 7), sessionsLeft: 8, free: true),
        const [],
        fixedNow,
      );

      expect(free.showPaidUntil, isFalse);
      expect(free.showSessions, isFalse);
    });

    test('hybrid shows both when both conditions are met', () {
      final visibility = StudentBalanceVisibility.of(
        _student(paidUntil: DateTime(2026, 11, 20), sessionsLeft: 4),
        const [],
        fixedNow,
      );

      expect(visibility.showPaidUntil, isTrue);
      expect(visibility.showSessions, isTrue);
    });
  });

  group('StudentHeaderDelegate', () {
    final fixedNow = DateTime(2026, 11, 8, 12, 0);
    final futurePaidUntil = DateTime(2026, 12, 7);
    final pastPaidUntil = DateTime(2026, 11, 1);

    const paidUntilLabel = '07.12.2026';
    const sessionsLabel = '8';

    testWidgets('shows both lines with status badges and balance badges', (
      tester,
    ) async {
      await _pumpHeader(
        tester,
        _student(
          statuses: const [StudentStatus.present, StudentStatus.paid],
          paidUntil: futurePaidUntil,
          sessionsLeft: 8,
        ),
        now: fixedNow,
      );

      // Line 1: Present, Paid
      expect(find.text(t.students.status.present), findsOneWidget);
      expect(find.text(t.students.status.paid), findsOneWidget);

      // Line 2: 8 sessions left, Paid until 07.12.2026
      expect(find.text(sessionsLabel), findsOneWidget);
      expect(find.text(paidUntilLabel), findsOneWidget);
    });

    testWidgets('hides sessions left when <= 0, shows paid until when valid', (
      tester,
    ) async {
      await _pumpHeader(
        tester,
        _student(
          statuses: const [StudentStatus.present, StudentStatus.paid],
          paidUntil: futurePaidUntil,
          sessionsLeft: 0,
        ),
        now: fixedNow,
      );

      expect(find.text(t.students.status.present), findsOneWidget);
      expect(find.text(t.students.status.paid), findsOneWidget);
      expect(find.text(paidUntilLabel), findsOneWidget);
      expect(find.textContaining('session'), findsNothing);
    });

    testWidgets('hides paid until when expired, shows sessions left when > 0', (
      tester,
    ) async {
      await _pumpHeader(
        tester,
        _student(
          statuses: const [StudentStatus.missing, StudentStatus.unpaid],
          paidUntil: pastPaidUntil,
          sessionsLeft: 8,
        ),
        now: fixedNow,
      );

      expect(find.text(t.students.status.missing), findsOneWidget);
      expect(find.text(t.students.status.unpaid), findsOneWidget);
      expect(find.text(sessionsLabel), findsOneWidget);
      expect(find.textContaining('Paid until'), findsNothing);
    });

    testWidgets('does not overflow when badges wrap on mobile screen', (
      tester,
    ) async {
      tester.view.physicalSize = const Size(360, 800);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final paidUntil = DateTime(2026, 11, 30);
      await _pumpHeader(
        tester,
        StudentSummaryResponse(
          id: 'student-333',
          fullName: '333',
          createdAt: DateTime(2025, 8, 22),
          free: false,
          statuses: const [StudentStatus.present, StudentStatus.paid],
          paidUntil: paidUntil,
          sessionsLeft: 23,
          phone: '+7 999 123-45-67',
        ),
        now: fixedNow,
      );

      expect(tester.takeException(), isNull);
      expect(find.text('333'), findsOneWidget);
      expect(find.text('+7 999 123-45-67'), findsOneWidget);
      expect(find.text(t.students.status.present), findsOneWidget);
      expect(find.text(t.students.status.paid), findsOneWidget);
      expect(find.text('23'), findsOneWidget);
      expect(find.text('30.11.2026'), findsOneWidget);
    });
  });
}
