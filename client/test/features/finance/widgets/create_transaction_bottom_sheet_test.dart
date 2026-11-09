import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:phone/features/finance/services/metric_service.dart';
import 'package:phone/features/finance/services/transaction_service.dart';
import 'package:phone/features/finance/widgets/create_transaction_bottom_sheet.dart';
import 'package:phone/features/finance/widgets/paid_from_field.dart';
import 'package:phone/features/finance/widgets/paid_until_field.dart';
import 'package:phone/features/finance/widgets/student_payment_mode_selector.dart';
import 'package:phone/features/students/services/student_service.dart';
import 'package:phone/generated/metric_controller/metric_controller_client.dart';
import 'package:phone/generated/models/create_student_payment_request.dart';
import 'package:phone/generated/models/metric_scope.dart';
import 'package:phone/generated/models/page_transaction_response.dart';
import 'package:phone/generated/models/payment_metric_response.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/transaction_type.dart';
import 'package:phone/generated/models/visit_with_lesson_response.dart';
import 'package:phone/generated/student_controller/student_controller_client.dart';
import 'package:phone/generated/transaction_controller/transaction_controller_client.dart';
import 'package:phone/i18n/strings.g.dart';

class _FakeStudentClient implements StudentControllerClient {
  new([this.students = const []]);

  final List<StudentSummaryResponse> students;
  final List<CreateStudentPaymentRequest> createdPayments = [];
  final List<String> createdPaymentStudentIds = [];
  List<VisitWithLessonResponse> visits = const [];
  List<TransactionResponse> payments = const [];

  @override
  Future<List<StudentSummaryResponse>> getStudentsSummary() async => students;

  @override
  Future<List<VisitWithLessonResponse>> getStudentVisits({
    required String studentId,
    DateTime? from,
    DateTime? to,
  }) async => visits;

  @override
  Future<List<TransactionResponse>> getStudentPayments({
    required String studentId,
  }) async => payments;

  @override
  Future<TransactionResponse> createStudentPayment({
    required String studentId,
    required CreateStudentPaymentRequest body,
  }) async {
    createdPaymentStudentIds.add(studentId);
    createdPayments.add(body);
    return TransactionResponse(
      id: 'tx-1',
      amount: body.amount,
      date: body.date,
      type: TransactionType.income,
      createdAt: DateTime(2025, 8, 22),
    );
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class _FakeTransactionClient implements TransactionControllerClient {
  @override
  Future<PageTransactionResponse> getPaginatedTransactions({
    int? page = 1,
    int? size = 20,
  }) async => const PageTransactionResponse(content: [], totalElements: 0);

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class _FakeMetricClient implements MetricControllerClient {
  @override
  Future<List<PaymentMetricResponse>> getPaymentStatistics({
    required DateTime startDate,
    required DateTime endDate,
    MetricScope? scope = MetricScope.month,
  }) async => const [];
}

final _student = StudentSummaryResponse(
  id: 'student-1',
  fullName: 'Anna Smirnova',
  createdAt: DateTime(2025, 8, 22),
  free: false,
  statuses: const [],
);

final _otherStudent = StudentSummaryResponse(
  id: 'student-2',
  fullName: 'Boris Petrov',
  createdAt: DateTime(2025, 8, 22),
  free: false,
  statuses: const [],
);

Widget _wrap({
  num? initialAmount,
  List<StudentSummaryResponse> students = const [],
  String? initialStudentId = 'student-1',
  String? initialStudentName,
  bool isStudentLocked = false,
  _FakeStudentClient? studentClient,
}) => ProviderScope(
  overrides: [
    studentServiceProvider.overrideWithValue(
      studentClient ?? _FakeStudentClient(students),
    ),
    transactionServiceProvider.overrideWithValue(_FakeTransactionClient()),
    metricServiceProvider.overrideWithValue(_FakeMetricClient()),
  ],
  child: TranslationProvider(
    child: MaterialApp(
      home: Scaffold(
        body: CreateTransactionBottomSheet(
          initialStudentId: initialStudentId,
          initialStudentName: initialStudentName,
          initialAmount: initialAmount,
          isStudentLocked: isStudentLocked,
        ),
      ),
    ),
  ),
);

Future<void> _pump(
  WidgetTester tester, {
  num? initialAmount,
  List<StudentSummaryResponse> students = const [],
  String? initialStudentId = 'student-1',
  String? initialStudentName,
  bool isStudentLocked = false,
  _FakeStudentClient? studentClient,
}) async {
  tester.view.physicalSize = const Size(1200, 2400);
  tester.view.devicePixelRatio = 1.0;
  addTearDown(tester.view.reset);

  await tester.pumpWidget(
    _wrap(
      initialAmount: initialAmount,
      students: students,
      initialStudentId: initialStudentId,
      initialStudentName: initialStudentName,
      isStudentLocked: isStudentLocked,
      studentClient: studentClient,
    ),
  );
  await tester.pumpAndSettle();
}

String _formatDate(DateTime date) {
  final day = date.day.toString().padLeft(2, '0');
  final month = date.month.toString().padLeft(2, '0');
  return '$day.$month.${date.year}';
}

String _digits(DateTime date) => _formatDate(date).replaceAll('.', '');

Finder _field(String label) => find.widgetWithText(TextField, label);

String _textOf(WidgetTester tester, String label) =>
    tester.widget<TextField>(_field(label)).controller!.text;

String get _dateLabel => '${t.finance.date}*';

String get _paidUntilLabel => '${t.finance.paidUntil}*';

void main() {
  setUpAll(() async {
    await initializeDateFormatting();
    LocaleSettings.setLocaleSync(AppLocale.en);
  });

  testWidgets('the date field is prefilled with today', (tester) async {
    await _pump(tester);

    expect(find.text(_dateLabel), findsOneWidget);
    expect(_textOf(tester, _dateLabel), _formatDate(DateTime.now()));
  });

  testWidgets('the paid until field is prefilled one month ahead', (
    tester,
  ) async {
    await _pump(tester);

    final now = DateTime.now();
    final expected = DateTime(now.year, now.month + 1, now.day);

    expect(find.text(_paidUntilLabel), findsOneWidget);
    expect(_textOf(tester, _paidUntilLabel), _formatDate(expected));
  });

  testWidgets('the date fields are plain text fields, not calendars', (
    tester,
  ) async {
    await _pump(tester);

    for (final label in [_dateLabel, _paidUntilLabel]) {
      final field = tester.widget<TextField>(_field(label));

      expect(field.readOnly, isFalse, reason: label);
      expect(field.enabled, isTrue, reason: label);
    }
  });

  testWidgets('a typed paid until date is masked as dd.MM.yyyy', (
    tester,
  ) async {
    await _pump(tester);

    final now = DateTime.now();
    final target = DateTime(now.year, now.month + 1);

    await tester.enterText(_field(_paidUntilLabel), _digits(target));
    await tester.pumpAndSettle();

    expect(_textOf(tester, _paidUntilLabel), _formatDate(target));
    expect(find.text(t.invalidDate), findsNothing);
  });

  testWidgets('a paid until date in the previous month is accepted', (
    tester,
  ) async {
    await _pump(tester, initialAmount: 1000);

    final now = DateTime.now();
    final previousMonthStart = DateTime(now.year, now.month - 1, 1);
    final previousMonth = DateTime(now.year, now.month - 1, 15);
    final earlierPaymentDate = DateTime(now.year, now.month, 1);

    await tester.enterText(
      _field(t.finance.paidFrom),
      _digits(previousMonthStart),
    );
    await tester.enterText(_field(_paidUntilLabel), _digits(previousMonth));
    await tester.pumpAndSettle();

    expect(_textOf(tester, _paidUntilLabel), _formatDate(previousMonth));
    expect(find.text(t.invalidDate), findsNothing);
    expect(
      tester
          .widget<TextButton>(find.widgetWithText(TextButton, t.create))
          .onPressed,
      isNotNull,
    );

    await tester.enterText(_field(_dateLabel), _digits(earlierPaymentDate));
    await tester.pumpAndSettle();

    expect(_textOf(tester, _paidUntilLabel), _formatDate(previousMonth));
    expect(find.text(t.invalidDate), findsNothing);
  });

  testWidgets('an impossible paid until date blocks saving', (tester) async {
    await _pump(tester, initialAmount: 1000);

    await tester.enterText(_field(_paidUntilLabel), '99999999');
    await tester.pumpAndSettle();

    expect(_textOf(tester, _paidUntilLabel), '99.99.9999');
    expect(find.text(t.invalidDate), findsOneWidget);
    expect(
      tester
          .widget<TextButton>(find.widgetWithText(TextButton, t.create))
          .onPressed,
      isNull,
    );
  });

  testWidgets('a date after today is rejected and blocks saving', (
    tester,
  ) async {
    await _pump(tester, initialAmount: 1000);

    final later = DateTime.now().add(const Duration(days: 30));

    await tester.enterText(_field(_dateLabel), _digits(later));
    await tester.pumpAndSettle();

    expect(find.text(t.invalidDate), findsOneWidget);
    expect(
      tester
          .widget<TextButton>(find.widgetWithText(TextButton, t.create))
          .onPressed,
      isNull,
    );
  });

  testWidgets('an amount of 0 blocks saving', (tester) async {
    await _pump(tester);

    await tester.enterText(_field('${t.finance.amount}*'), '0');
    await tester.pumpAndSettle();

    expect(
      tester
          .widget<TextButton>(find.widgetWithText(TextButton, t.create))
          .onPressed,
      isNull,
    );
  });

  testWidgets('a positive amount activates the create button', (tester) async {
    await _pump(tester);

    await tester.enterText(_field('${t.finance.amount}*'), '100');
    await tester.pumpAndSettle();

    expect(
      tester
          .widget<TextButton>(find.widgetWithText(TextButton, t.create))
          .onPressed,
      isNotNull,
    );
  });

  testWidgets('the student picker offers the other students by default', (
    tester,
  ) async {
    await _pump(tester, students: [_student, _otherStudent]);

    expect(find.text(_student.fullName), findsOneWidget);
    expect(find.text(_otherStudent.fullName), findsNothing);

    await tester.tap(find.text(_student.fullName));
    await tester.pumpAndSettle();

    expect(find.text(_otherStudent.fullName), findsOneWidget);
  });

  testWidgets('a locked sheet keeps the student of the page', (tester) async {
    await _pump(
      tester,
      students: [_student, _otherStudent],
      initialStudentName: _student.fullName,
      isStudentLocked: true,
    );

    expect(find.text(_student.fullName), findsOneWidget);

    await tester.tap(find.text(_student.fullName));
    await tester.pumpAndSettle();

    expect(find.text(_otherStudent.fullName), findsNothing);
    expect(find.text(t.students.noStudent), findsNothing);
  });

  testWidgets('a locked sheet shows the student before the list is loaded', (
    tester,
  ) async {
    await _pump(
      tester,
      initialStudentName: _student.fullName,
      isStudentLocked: true,
    );

    expect(find.text(_student.fullName), findsOneWidget);
  });

  testWidgets(
    'payment mode selector is shown only when an income student is selected',
    (tester) async {
      await _pump(tester, initialStudentId: null, students: [_student]);

      expect(find.byType(StudentPaymentModeSelector), findsNothing);
      expect(_field(_paidUntilLabel), findsNothing);

      await tester.tap(find.text(t.students.noStudent));
      await tester.pumpAndSettle();
      await tester.tap(find.text(_student.fullName));
      await tester.pumpAndSettle();

      expect(find.byType(StudentPaymentModeSelector), findsOneWidget);
      expect(_field(_paidUntilLabel), findsOneWidget);

      await tester.tap(find.text(t.finance.typeExpense));
      await tester.pumpAndSettle();

      expect(find.byType(StudentPaymentModeSelector), findsNothing);
    },
  );

  testWidgets(
    'switching between Paid until and Sessions count swaps the input field and enforces digits only',
    (tester) async {
      await _pump(tester, initialAmount: 1000);

      final sessionsLabel = '${t.finance.sessionsCount}*';

      expect(_field(_paidUntilLabel), findsOneWidget);
      expect(_field(sessionsLabel), findsNothing);

      await tester.tap(find.text(t.finance.sessionsCount));
      await tester.pumpAndSettle();

      expect(_field(_paidUntilLabel), findsNothing);
      expect(_field(sessionsLabel), findsOneWidget);
      expect(
        tester
            .widget<TextButton>(find.widgetWithText(TextButton, t.create))
            .onPressed,
        isNull,
      );

      await tester.enterText(_field(sessionsLabel), '0');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<TextButton>(find.widgetWithText(TextButton, t.create))
            .onPressed,
        isNull,
      );

      await tester.enterText(_field(sessionsLabel), 'a-12.5b');
      await tester.pumpAndSettle();

      expect(_textOf(tester, sessionsLabel), '125');
      expect(
        tester
            .widget<TextButton>(find.widgetWithText(TextButton, t.create))
            .onPressed,
        isNotNull,
      );

      await tester.tap(find.text(t.finance.paidUntil));
      await tester.pumpAndSettle();

      expect(_field(_paidUntilLabel), findsOneWidget);
      expect(_field(sessionsLabel), findsNothing);
    },
  );

  testWidgets(
    'submitting in Sessions count mode sends sessionsAdded without paidUntil',
    (tester) async {
      final client = _FakeStudentClient([_student]);
      await _pump(tester, initialAmount: 2500, studentClient: client);

      await tester.tap(find.text(t.finance.sessionsCount));
      await tester.pumpAndSettle();

      await tester.enterText(_field('${t.finance.sessionsCount}*'), '8');
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(TextButton, t.create));
      await tester.pumpAndSettle();

      expect(client.createdPaymentStudentIds, ['student-1']);
      expect(client.createdPayments, hasLength(1));
      expect(client.createdPayments.single.amount, 2500);
      expect(client.createdPayments.single.paidUntil, isNull);
      expect(client.createdPayments.single.sessionsAdded, 8);
    },
  );

  testWidgets(
    'submitting in Paid until mode sends paidUntil without sessionsAdded',
    (tester) async {
      final client = _FakeStudentClient([_student]);
      await _pump(tester, initialAmount: 2500, studentClient: client);

      await tester.tap(find.widgetWithText(TextButton, t.create));
      await tester.pumpAndSettle();

      final now = DateTime.now();
      final expectedPaidUntil = DateTime(now.year, now.month + 1, now.day);

      expect(client.createdPaymentStudentIds, ['student-1']);
      expect(client.createdPayments, hasLength(1));
      expect(client.createdPayments.single.amount, 2500);
      expect(client.createdPayments.single.paidUntil, expectedPaidUntil);
      expect(client.createdPayments.single.sessionsAdded, isNull);
    },
  );

  testWidgets(
    'renders Russian selector and sessions field cleanly on narrow screens',
    (tester) async {
      await tester.runAsync(() => LocaleSettings.setLocale(AppLocale.ru));
      addTearDown(() => LocaleSettings.setLocaleSync(AppLocale.en));

      tester.view.physicalSize = const Size(360, 800);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.reset);

      await tester.pumpWidget(_wrap(initialAmount: 1500));
      await tester.pumpAndSettle();

      expect(find.text('Оплачено до'), findsOneWidget);
      expect(find.text('Количество занятий'), findsOneWidget);

      await tester.tap(find.text('Количество занятий'));
      await tester.pumpAndSettle();

      expect(_field('Количество занятий*'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('paid from and paid until fields are displayed in the same row', (
    tester,
  ) async {
    await _pump(tester, initialAmount: 2000);

    expect(find.byType(PaidFromField), findsOneWidget);
    expect(find.byType(PaidUntilField), findsOneWidget);

    final rowFinder = find.ancestor(
      of: find.byType(PaidFromField),
      matching: find.byType(Row),
    );
    expect(rowFinder, findsOneWidget);

    final untilInSameRow = find.descendant(
      of: rowFinder,
      matching: find.byType(PaidUntilField),
    );
    expect(untilInSameRow, findsOneWidget);
  });

  testWidgets(
    'prefills paid from with today and paid until with today + 1 month',
    (tester) async {
      await _pump(tester, initialAmount: 3000, initialStudentId: 'student-1');

      final now = DateTime.now();
      final expectedPaidUntil = DateTime(now.year, now.month + 1, now.day);
      expect(_textOf(tester, t.finance.paidFrom), _formatDate(now));
      expect(_textOf(tester, _paidUntilLabel), _formatDate(expectedPaidUntil));
    },
  );

  testWidgets('paid from field does not have a clear button', (tester) async {
    await _pump(tester, initialAmount: 3000, initialStudentId: 'student-1');

    final paidFromFinder = find.byType(PaidFromField);
    expect(
      find.descendant(of: paidFromFinder, matching: find.byIcon(Icons.clear)),
      findsNothing,
    );
  });

  testWidgets('entering paidFrom does not auto-update paidUntil', (
    tester,
  ) async {
    await _pump(tester, initialAmount: 3500);

    final initialPaidUntil = _textOf(tester, _paidUntilLabel);
    await tester.enterText(_field(t.finance.paidFrom), '10042025');
    await tester.pump();

    expect(_textOf(tester, _paidUntilLabel), initialPaidUntil);
  });

  testWidgets('submitting payment sends both paidFrom and paidUntil', (
    tester,
  ) async {
    final client = _FakeStudentClient([_student]);
    await _pump(tester, initialAmount: 3500, studentClient: client);

    await tester.enterText(_field(t.finance.paidFrom), '01082025');
    await tester.enterText(_field(_paidUntilLabel), '31082025');
    await tester.pump();

    await tester.tap(find.text(t.create));
    await tester.pumpAndSettle();

    expect(client.createdPayments, hasLength(1));
    expect(client.createdPayments.single.paidFrom, DateTime(2025, 8, 1));
    expect(client.createdPayments.single.paidUntil, DateTime(2025, 8, 31));
  });

  testWidgets(
    'when paidFrom is after paidUntil shows error under both date fields and blocks saving',
    (tester) async {
      final client = _FakeStudentClient([_student]);
      await _pump(tester, initialAmount: 3500, studentClient: client);

      await tester.enterText(_field(t.finance.paidFrom), '15092025');
      await tester.enterText(_field(_paidUntilLabel), '01092025');
      await tester.pump();

      final paidFromFinder = find.byType(PaidFromField);
      final paidUntilFinder = find.byType(PaidUntilField);
      expect(
        find.descendant(of: paidFromFinder, matching: find.text(t.invalidDate)),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: paidUntilFinder,
          matching: find.text(t.invalidDate),
        ),
        findsOneWidget,
      );

      await tester.tap(find.text(t.create));
      await tester.pumpAndSettle();

      expect(client.createdPayments, isEmpty);

      // Once paidUntil is corrected to be after paidFrom, error under both fields is cleared
      await tester.enterText(_field(_paidUntilLabel), '20092025');
      await tester.pump();

      expect(
        find.descendant(of: paidFromFinder, matching: find.text(t.invalidDate)),
        findsNothing,
      );
      expect(
        find.descendant(
          of: paidUntilFinder,
          matching: find.text(t.invalidDate),
        ),
        findsNothing,
      );
    },
  );
}
