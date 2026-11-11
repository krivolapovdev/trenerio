import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:phone/features/finance/services/metric_service.dart';
import 'package:phone/features/finance/services/transaction_service.dart';
import 'package:phone/features/groups/pages/group_payments_page.dart';
import 'package:phone/features/groups/services/group_service.dart';
import 'package:phone/features/groups/widgets/student_payment_bottom_sheet.dart';
import 'package:phone/features/students/services/student_service.dart';
import 'package:phone/generated/export.dart';
import 'package:phone/i18n/strings.g.dart';

final GroupSummaryResponse _testGroup = GroupSummaryResponse(
  id: 'group-1',
  name: 'Swimming Group',
  createdAt: DateTime(2026, 1, 1),
  defaultPrice: 3000,
);

final GroupStudentSummaryResponse _student1 = GroupStudentSummaryResponse(
  id: 'student-1',
  fullName: 'Alice Johnson',
  createdAt: DateTime(2026, 1, 1),
  free: false,
  statuses: const [StudentStatus.unpaid],
);

final GroupStudentSummaryResponse _student2 = GroupStudentSummaryResponse(
  id: 'student-2',
  fullName: 'Bob Smith',
  createdAt: DateTime(2026, 1, 1),
  free: false,
  statuses: const [StudentStatus.unpaid],
);

final GroupStudentSummaryResponse _studentFree = GroupStudentSummaryResponse(
  id: 'student-3',
  fullName: 'Charlie Free',
  createdAt: DateTime(2026, 1, 1),
  free: true,
  statuses: const [StudentStatus.free],
);

class _FakeGroupClient implements GroupControllerClient {
  final List<GroupStudentSummaryResponse> students;

  new({this.students = const []});

  @override
  Future<List<GroupStudentSummaryResponse>> getGroupStudents({
    required String groupId,
  }) async => students;

  @override
  Future<List<GroupSummaryResponse>> getAllGroupsSummary() async => [
    _testGroup,
  ];

  @override
  Future<GroupResponse> createGroup({required CreateGroupRequest body}) =>
      throw UnimplementedError();

  @override
  Future<void> deleteGroup({required String groupId}) =>
      throw UnimplementedError();

  @override
  Future<List<LessonResponse>> getGroupLessons({
    required String groupId,
    required DateTime from,
    required DateTime to,
  }) async => const [];

  @override
  Future<GroupReportResponse> getGroupReport({
    required String groupId,
    required int year,
    required int month,
  }) => throw UnimplementedError();

  @override
  Future<GroupResponse> updateGroup({
    required String groupId,
    required Map<String, dynamic> body,
  }) => throw UnimplementedError();
}

class _FakeStudentClient implements StudentControllerClient {
  final List<CreateStudentPaymentRequest> createdPayments = [];
  final List<String> createdPaymentStudentIds = [];

  @override
  Future<TransactionResponse> createStudentPayment({
    required String studentId,
    required CreateStudentPaymentRequest body,
  }) async {
    createdPaymentStudentIds.add(studentId);
    createdPayments.add(body);
    return TransactionResponse(
      amount: body.amount,
      createdAt: DateTime.now(),
      date: body.date,
      id: 'tx-1',
      type: TransactionType.income,
    );
  }

  @override
  Future<List<StudentSummaryResponse>> getStudentsSummary() async => const [];

  @override
  Future<StudentResponse> createStudent({required CreateStudentRequest body}) =>
      throw UnimplementedError();

  @override
  Future<void> deleteStudent({required String studentId}) =>
      throw UnimplementedError();

  @override
  Future<List<TransactionResponse>> getStudentPayments({
    required String studentId,
  }) async => const [];

  @override
  Future<List<VisitWithLessonResponse>> getStudentVisits({
    required String studentId,
    DateTime? from,
    DateTime? to,
  }) async => const [];

  @override
  Future<StudentResponse> updateStudent({
    required String studentId,
    required Map<String, dynamic> body,
  }) => throw UnimplementedError();
}

class _FakeTransactionClient implements TransactionControllerClient {
  @override
  Future<PageTransactionResponse> getPaginatedTransactions({
    int? page = 1,
    int? size = 20,
  }) async => const PageTransactionResponse(
    content: [],
    pageable: null,
    totalElements: 0,
    totalPages: 0,
  );

  @override
  Future<TransactionResponse> createTransaction({
    required CreateTransactionRequest body,
  }) => throw UnimplementedError();

  @override
  Future<void> deleteTransaction({required String transactionId}) =>
      throw UnimplementedError();

  @override
  Future<TransactionResponse> updateTransaction({
    required String transactionId,
    required Map<String, dynamic> body,
  }) => throw UnimplementedError();
}

class _FakeMetricClient implements MetricControllerClient {
  @override
  Future<List<PaymentMetricResponse>> getPaymentStatistics({
    required DateTime startDate,
    required DateTime endDate,
    MetricScope? scope = MetricScope.month,
  }) async => const [];
}

Widget _wrap({
  required Widget child,
  required _FakeGroupClient groupClient,
  required _FakeStudentClient studentClient,
}) => ProviderScope(
  overrides: [
    groupServiceProvider.overrideWithValue(groupClient),
    studentServiceProvider.overrideWithValue(studentClient),
    transactionServiceProvider.overrideWithValue(_FakeTransactionClient()),
    metricServiceProvider.overrideWithValue(_FakeMetricClient()),
  ],
  child: TranslationProvider(child: MaterialApp(home: child)),
);

void main() {
  setUpAll(() => LocaleSettings.setLocaleSync(AppLocale.en));

  testWidgets(
    'renders group payments page with pre-filled default price and students',
    (tester) async {
      final groupClient = _FakeGroupClient(
        students: [_student1, _student2, _studentFree],
      );
      final studentClient = _FakeStudentClient();

      await tester.pumpWidget(
        _wrap(
          child: GroupPaymentsPage(group: _testGroup),
          groupClient: groupClient,
          studentClient: studentClient,
        ),
      );
      await tester.pumpAndSettle();

      expect(
        find.text(t.groups.groupPaymentsTitle(name: 'Swimming Group')),
        findsOneWidget,
      );
      expect(find.text('3000'), findsOneWidget); // defaultPrice prefilled
      expect(find.text('Alice Johnson'), findsOneWidget);
      expect(find.text('Bob Smith'), findsOneWidget);
      expect(find.text('Charlie Free'), findsOneWidget);

      // Alice and Bob are checked, Charlie (free) is unchecked
      final checkboxes = tester
          .widgetList<Checkbox>(find.byType(Checkbox))
          .map((c) => c.value)
          .toList();
      expect(checkboxes, [true, true, false]);
    },
  );

  testWidgets('toggle student checkbox toggles selection', (tester) async {
    final groupClient = _FakeGroupClient(students: [_student1, _student2]);
    final studentClient = _FakeStudentClient();

    await tester.pumpWidget(
      _wrap(
        child: GroupPaymentsPage(group: _testGroup),
        groupClient: groupClient,
        studentClient: studentClient,
      ),
    );
    await tester.pumpAndSettle();

    // Tap on Alice Johnson
    await tester.tap(find.text('Alice Johnson'));
    await tester.pumpAndSettle();

    final checkboxes = tester
        .widgetList<Checkbox>(find.byType(Checkbox))
        .map((c) => c.value)
        .toList();
    expect(checkboxes, [false, true]);
  });

  testWidgets('toggle all selects and deselects all paying students', (
    tester,
  ) async {
    final groupClient = _FakeGroupClient(
      students: [_student1, _student2, _studentFree],
    );
    final studentClient = _FakeStudentClient();

    await tester.pumpWidget(
      _wrap(
        child: GroupPaymentsPage(group: _testGroup),
        groupClient: groupClient,
        studentClient: studentClient,
      ),
    );
    await tester.pumpAndSettle();

    // Initially Alice and Bob are selected -> shows "Deselect all"
    expect(find.text(t.lessons.deselectAll), findsOneWidget);

    await tester.tap(find.text(t.lessons.deselectAll));
    await tester.pumpAndSettle();

    var checkboxes = tester
        .widgetList<Checkbox>(find.byType(Checkbox))
        .map((c) => c.value)
        .toList();
    expect(checkboxes, [false, false, false]);

    // Now shows "Select all"
    expect(find.text(t.lessons.selectAll), findsOneWidget);

    await tester.tap(find.text(t.lessons.selectAll));
    await tester.pumpAndSettle();

    checkboxes = tester
        .widgetList<Checkbox>(find.byType(Checkbox))
        .map((c) => c.value)
        .toList();
    expect(checkboxes, [true, true, false]); // Charlie Free is not selected
  });

  testWidgets(
    'submitting creates payments for selected students and pops page',
    (tester) async {
      final groupClient = _FakeGroupClient(
        students: [_student1, _student2, _studentFree],
      );
      final studentClient = _FakeStudentClient();

      await tester.pumpWidget(
        _wrap(
          child: Builder(
            builder: (context) => TextButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => GroupPaymentsPage(group: _testGroup),
                ),
              ),
              child: const Text('Open Payments'),
            ),
          ),
          groupClient: groupClient,
          studentClient: studentClient,
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Open Payments'));
      await tester.pumpAndSettle();

      expect(find.byType(GroupPaymentsPage), findsOneWidget);

      // Tap create
      await tester.tap(find.text(t.create));
      await tester.pumpAndSettle();

      // Alice and Bob payments should be created
      expect(studentClient.createdPaymentStudentIds, [
        'student-1',
        'student-2',
      ]);
      expect(studentClient.createdPayments.length, 2);
      expect(studentClient.createdPayments[0].amount, 3000);
      expect(studentClient.createdPayments[1].amount, 3000);

      // Page should be popped
      expect(find.byType(GroupPaymentsPage), findsNothing);
    },
  );

  testWidgets('can customize individual student amount via bottom sheet', (
    tester,
  ) async {
    final groupClient = _FakeGroupClient(students: [_student1]);
    final studentClient = _FakeStudentClient();

    await tester.pumpWidget(
      _wrap(
        child: GroupPaymentsPage(group: _testGroup),
        groupClient: groupClient,
        studentClient: studentClient,
      ),
    );
    await tester.pumpAndSettle();

    // Tap more button on Alice Johnson
    await tester.tap(find.byIcon(Icons.more_vert));
    await tester.pumpAndSettle();

    expect(find.byType(StudentPaymentBottomSheet), findsOneWidget);

    // Enter custom amount
    final amountInput = find.descendant(
      of: find.byType(StudentPaymentBottomSheet),
      matching: find.widgetWithText(TextField, '${t.finance.amount}*'),
    );
    await tester.enterText(amountInput, '2500');
    await tester.pumpAndSettle();

    // Save in bottom sheet
    final saveButton = find.descendant(
      of: find.byType(StudentPaymentBottomSheet),
      matching: find.text(t.create),
    );
    await tester.tap(saveButton);
    await tester.pumpAndSettle();

    // Custom amount badge/text should be visible on the tile
    expect(find.textContaining(t.groups.customAmount), findsOneWidget);

    // Submit group payments
    await tester.tap(
      find.descendant(of: find.byType(AppBar), matching: find.text(t.create)),
    );
    await tester.pumpAndSettle();

    expect(studentClient.createdPayments.single.amount, 2500);
  });

  testWidgets(
    'clearing paid from date card sends payments with null paidFrom',
    (tester) async {
      final groupClient = _FakeGroupClient(students: [_student1]);
      final studentClient = _FakeStudentClient();

      await tester.pumpWidget(
        _wrap(
          child: GroupPaymentsPage(group: _testGroup),
          groupClient: groupClient,
          studentClient: studentClient,
        ),
      );
      await tester.pumpAndSettle();

      // Tap clear on paid from card
      final clearPaidFrom = find.descendant(
        of: find.byWidgetPredicate(
          (w) =>
              w is InputDecorator &&
              w.decoration.labelText == t.finance.paidFrom,
        ),
        matching: find.byIcon(Icons.clear),
      );
      expect(clearPaidFrom, findsOneWidget);
      await tester.tap(clearPaidFrom);
      await tester.pumpAndSettle();

      // Submit group payments
      await tester.tap(
        find.descendant(of: find.byType(AppBar), matching: find.text(t.create)),
      );
      await tester.pumpAndSettle();

      expect(studentClient.createdPayments.single.amount, 3000);
      expect(studentClient.createdPayments.single.paidFrom, isNull);
      expect(studentClient.createdPayments.single.paidUntil, isNotNull);
    },
  );

  testWidgets(
    'switching to sessions mode and entering count sends sessionsAdded',
    (tester) async {
      final groupClient = _FakeGroupClient(students: [_student1]);
      final studentClient = _FakeStudentClient();

      await tester.pumpWidget(
        _wrap(
          child: GroupPaymentsPage(group: _testGroup),
          groupClient: groupClient,
          studentClient: studentClient,
        ),
      );
      await tester.pumpAndSettle();

      // Switch to sessions count
      await tester.tap(find.text(t.finance.sessionsCount));
      await tester.pumpAndSettle();

      // Type sessions count
      final sessionsInput = find.widgetWithText(
        TextField,
        '${t.finance.sessionsCount}*',
      );
      expect(sessionsInput, findsOneWidget);
      await tester.enterText(sessionsInput, '8');
      await tester.pumpAndSettle();

      // Submit group payments
      await tester.tap(
        find.descendant(of: find.byType(AppBar), matching: find.text(t.create)),
      );
      await tester.pumpAndSettle();

      expect(studentClient.createdPayments.single.amount, 3000);
      expect(studentClient.createdPayments.single.sessionsAdded, 8);
      expect(studentClient.createdPayments.single.paidUntil, isNull);
      expect(studentClient.createdPayments.single.paidFrom, isNull);
    },
  );

  testWidgets(
    'can customize individual student sessions count via bottom sheet',
    (tester) async {
      final groupClient = _FakeGroupClient(students: [_student1]);
      final studentClient = _FakeStudentClient();

      await tester.pumpWidget(
        _wrap(
          child: GroupPaymentsPage(group: _testGroup),
          groupClient: groupClient,
          studentClient: studentClient,
        ),
      );
      await tester.pumpAndSettle();

      // Tap more button on Alice Johnson
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();

      expect(find.byType(StudentPaymentBottomSheet), findsOneWidget);

      // In bottom sheet, switch to sessions count mode
      final sessionsModeInSheet = find.descendant(
        of: find.byType(StudentPaymentBottomSheet),
        matching: find.text(t.finance.sessionsCount),
      );
      await tester.tap(sessionsModeInSheet);
      await tester.pumpAndSettle();

      final sheetSessionsInput = find.descendant(
        of: find.byType(StudentPaymentBottomSheet),
        matching: find.widgetWithText(TextField, '${t.finance.sessionsCount}*'),
      );
      await tester.enterText(sheetSessionsInput, '12');
      await tester.pumpAndSettle();

      // Save in bottom sheet
      final saveButton = find.descendant(
        of: find.byType(StudentPaymentBottomSheet),
        matching: find.text(t.create),
      );
      await tester.tap(saveButton);
      await tester.pumpAndSettle();

      // Custom sessions should be visible in subtitle
      expect(
        find.textContaining('${t.finance.sessionsCount}: 12'),
        findsOneWidget,
      );

      // Submit group payments
      await tester.tap(
        find.descendant(of: find.byType(AppBar), matching: find.text(t.create)),
      );
      await tester.pumpAndSettle();

      expect(studentClient.createdPayments.single.sessionsAdded, 12);
      expect(studentClient.createdPayments.single.paidUntil, isNull);
    },
  );
}
