import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:phone/features/finance/services/metric_service.dart';
import 'package:phone/features/finance/services/transaction_service.dart';
import 'package:phone/features/finance/widgets/payments/student_payment_bottom_sheet.dart';
import 'package:phone/features/groups/services/group_service.dart';
import 'package:phone/features/students/pages/student_payments_page.dart';
import 'package:phone/features/students/services/student_service.dart';
import 'package:phone/generated/export.dart';
import 'package:phone/i18n/strings.g.dart';

final StudentSummaryResponse _student1 = StudentSummaryResponse(
  id: 'student-1',
  fullName: 'Alice Johnson',
  createdAt: DateTime(2026, 1, 1),
  free: false,
  statuses: const [StudentStatus.unpaid],
);

final StudentSummaryResponse _student2 = StudentSummaryResponse(
  id: 'student-2',
  fullName: 'Bob Smith',
  createdAt: DateTime(2026, 1, 1),
  free: false,
  statuses: const [StudentStatus.unpaid],
);

final StudentSummaryResponse _studentFree = StudentSummaryResponse(
  id: 'student-3',
  fullName: 'Charlie Free',
  createdAt: DateTime(2026, 1, 1),
  free: true,
  statuses: const [StudentStatus.free],
);

class _FakeStudentClient implements StudentControllerClient {
  final List<StudentSummaryResponse> students;
  final List<CreateStudentPaymentRequest> createdPayments = [];
  final List<String> createdPaymentStudentIds = [];

  new({this.students = const []});

  @override
  Future<List<StudentSummaryResponse>> getStudentsSummary() async => students;

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

class _FakeGroupClient implements GroupControllerClient {
  @override
  Future<List<GroupSummaryResponse>> getAllGroupsSummary() async => const [];

  @override
  Future<List<GroupStudentSummaryResponse>> getGroupStudents({
    required String groupId,
  }) async => const [];

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
  required _FakeStudentClient studentClient,
}) => ProviderScope(
  overrides: [
    studentServiceProvider.overrideWithValue(studentClient),
    groupServiceProvider.overrideWithValue(_FakeGroupClient()),
    transactionServiceProvider.overrideWithValue(_FakeTransactionClient()),
    metricServiceProvider.overrideWithValue(_FakeMetricClient()),
  ],
  child: TranslationProvider(child: MaterialApp(home: child)),
);

void main() {
  setUpAll(() => LocaleSettings.setLocaleSync(AppLocale.en));

  testWidgets('renders student payments page with students', (tester) async {
    final studentClient = _FakeStudentClient(
      students: [_student1, _student2, _studentFree],
    );

    await tester.pumpWidget(
      _wrap(child: const StudentPaymentsPage(), studentClient: studentClient),
    );
    await tester.pumpAndSettle();

    expect(find.text(t.students.payments), findsOneWidget);
    expect(find.text('Alice Johnson'), findsOneWidget);
    expect(find.text('Bob Smith'), findsOneWidget);
    expect(find.text('Charlie Free'), findsOneWidget);

    final checkboxes = tester
        .widgetList<Checkbox>(find.byType(Checkbox))
        .map((c) => c.value)
        .toList();
    expect(checkboxes, [true, true, false]);
  });

  testWidgets('toggle student checkbox toggles selection', (tester) async {
    final studentClient = _FakeStudentClient(students: [_student1, _student2]);

    await tester.pumpWidget(
      _wrap(child: const StudentPaymentsPage(), studentClient: studentClient),
    );
    await tester.pumpAndSettle();

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
    final studentClient = _FakeStudentClient(
      students: [_student1, _student2, _studentFree],
    );

    await tester.pumpWidget(
      _wrap(child: const StudentPaymentsPage(), studentClient: studentClient),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text(t.lessons.deselectAll));
    await tester.pumpAndSettle();

    var checkboxes = tester
        .widgetList<Checkbox>(find.byType(Checkbox))
        .map((c) => c.value)
        .toList();
    expect(checkboxes, [false, false, false]);

    await tester.tap(find.text(t.lessons.selectAll));
    await tester.pumpAndSettle();

    checkboxes = tester
        .widgetList<Checkbox>(find.byType(Checkbox))
        .map((c) => c.value)
        .toList();
    expect(checkboxes, [true, true, false]);
  });

  testWidgets(
    'submitting creates payments for selected students and pops page',
    (tester) async {
      final studentClient = _FakeStudentClient(
        students: [_student1, _student2],
      );

      await tester.pumpWidget(
        _wrap(child: const StudentPaymentsPage(), studentClient: studentClient),
      );
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).first, '2500');
      await tester.pumpAndSettle();

      await tester.tap(find.text('Alice Johnson'));
      await tester.pumpAndSettle();

      await tester.tap(find.text(t.create));
      await tester.pumpAndSettle();

      expect(studentClient.createdPayments.length, 1);
      expect(studentClient.createdPaymentStudentIds, ['student-2']);
      expect(studentClient.createdPayments.first.amount, 2500);
    },
  );

  testWidgets('can customize individual student amount via bottom sheet', (
    tester,
  ) async {
    final studentClient = _FakeStudentClient(students: [_student1, _student2]);

    await tester.pumpWidget(
      _wrap(child: const StudentPaymentsPage(), studentClient: studentClient),
    );
    await tester.pumpAndSettle();

    await tester.enterText(find.byType(TextField).first, '2000');
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.more_vert).first);
    await tester.pumpAndSettle();

    expect(find.byType(StudentPaymentBottomSheet), findsOneWidget);

    final bottomSheetAmountField = find.widgetWithText(
      TextField,
      '${t.finance.amount}*',
    );
    await tester.enterText(bottomSheetAmountField, '4500');
    await tester.pumpAndSettle();

    final saveFinder = find.descendant(
      of: find.byType(StudentPaymentBottomSheet),
      matching: find.text(t.create),
    );
    await tester.tap(saveFinder);
    await tester.pumpAndSettle();

    expect(find.textContaining(t.groups.customAmount), findsOneWidget);

    await tester.tap(
      find.descendant(of: find.byType(AppBar), matching: find.text(t.create)),
    );
    await tester.pumpAndSettle();

    expect(studentClient.createdPayments.length, 2);
    final alicePayment =
        studentClient.createdPayments[studentClient.createdPaymentStudentIds
            .indexOf('student-1')];
    final bobPayment =
        studentClient.createdPayments[studentClient.createdPaymentStudentIds
            .indexOf('student-2')];

    expect(alicePayment.amount, 4500);
    expect(bobPayment.amount, 2000);
  });

  testWidgets(
    'switching to sessions mode and entering count sends sessionsAdded',
    (tester) async {
      final studentClient = _FakeStudentClient(students: [_student1]);

      await tester.pumpWidget(
        _wrap(child: const StudentPaymentsPage(), studentClient: studentClient),
      );
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).first, '1500');
      await tester.pumpAndSettle();

      await tester.tap(find.text(t.finance.sessionsCount));
      await tester.pumpAndSettle();

      final sessionsField = find.widgetWithText(
        TextField,
        '${t.finance.sessionsCount}*',
      );
      expect(sessionsField, findsOneWidget);
      await tester.enterText(sessionsField, '8');
      await tester.pumpAndSettle();

      await tester.tap(
        find.descendant(of: find.byType(AppBar), matching: find.text(t.create)),
      );
      await tester.pumpAndSettle();

      expect(studentClient.createdPayments.length, 1);
      expect(studentClient.createdPayments.first.amount, 1500);
      expect(studentClient.createdPayments.first.sessionsAdded, 8);
      expect(studentClient.createdPayments.first.paidUntil, isNull);
    },
  );
}
