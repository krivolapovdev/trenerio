import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:phone/features/finance/services/metric_service.dart';
import 'package:phone/features/finance/services/transaction_service.dart';
import 'package:phone/features/students/controllers/create_student_payment_controller.dart';
import 'package:phone/features/students/services/student_service.dart';
import 'package:phone/generated/export.dart';

class _FakeStudentClient implements StudentControllerClient {
  final List<CreateStudentPaymentRequest> createdPayments = [];
  final List<String> createdStudentIds = [];

  @override
  Future<TransactionResponse> createStudentPayment({
    required String studentId,
    required CreateStudentPaymentRequest body,
  }) async {
    createdStudentIds.add(studentId);
    createdPayments.add(body);
    return TransactionResponse(
      id: 'tx-1',
      amount: body.amount,
      date: body.date,
      type: TransactionType.income,
      createdAt: DateTime(2026, 10, 30),
    );
  }

  @override
  Future<List<StudentSummaryResponse>> getStudentsSummary() async => const [];

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

void main() {
  late _FakeStudentClient studentClient;
  late ProviderContainer container;

  setUp(() {
    studentClient = _FakeStudentClient();
    container = ProviderContainer(
      overrides: [
        studentServiceProvider.overrideWithValue(studentClient),
        transactionServiceProvider.overrideWithValue(_FakeTransactionClient()),
        metricServiceProvider.overrideWithValue(_FakeMetricClient()),
      ],
    );
    addTearDown(container.dispose);
  });

  test('saves a date-based student payment with paidUntil', () async {
    final controller = container.read(
      createStudentPaymentControllerProvider.notifier,
    );
    final date = DateTime(2026, 10, 1);
    final paidUntil = DateTime(2026, 11, 1);

    final ok = await controller.saveStudentPayment(
      studentId: 'student-1',
      amount: 3500,
      date: date,
      paidUntil: paidUntil,
    );

    expect(ok, isTrue);
    expect(studentClient.createdStudentIds, ['student-1']);
    expect(studentClient.createdPayments, hasLength(1));
    expect(studentClient.createdPayments.single.amount, 3500);
    expect(studentClient.createdPayments.single.date, date);
    expect(studentClient.createdPayments.single.paidUntil, paidUntil);
    expect(studentClient.createdPayments.single.sessionsAdded, isNull);
  });

  test('saves a session-based student payment with sessionsAdded', () async {
    final controller = container.read(
      createStudentPaymentControllerProvider.notifier,
    );
    final date = DateTime(2026, 10, 1);

    final ok = await controller.saveStudentPayment(
      studentId: 'student-2',
      amount: 5000,
      date: date,
      sessionsAdded: 12,
    );

    expect(ok, isTrue);
    expect(studentClient.createdStudentIds, ['student-2']);
    expect(studentClient.createdPayments, hasLength(1));
    expect(studentClient.createdPayments.single.amount, 5000);
    expect(studentClient.createdPayments.single.date, date);
    expect(studentClient.createdPayments.single.paidUntil, isNull);
    expect(studentClient.createdPayments.single.sessionsAdded, 12);
  });

  test('rejects student payment when both paidUntil and sessionsAdded are missing or non-positive', () async {
    final controller = container.read(
      createStudentPaymentControllerProvider.notifier,
    );

    final okMissing = await controller.saveStudentPayment(
      studentId: 'student-1',
      amount: 3500,
      date: DateTime(2026, 10, 1),
    );
    expect(okMissing, isFalse);
    expect(
      container.read(createStudentPaymentControllerProvider).hasError,
      isTrue,
    );
    expect(studentClient.createdPayments, isEmpty);

    final okZero = await controller.saveStudentPayment(
      studentId: 'student-1',
      amount: 3500,
      date: DateTime(2026, 10, 1),
      sessionsAdded: 0,
    );
    expect(okZero, isFalse);
    expect(studentClient.createdPayments, isEmpty);
  });
}
