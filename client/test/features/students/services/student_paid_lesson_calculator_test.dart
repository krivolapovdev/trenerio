import 'package:flutter_test/flutter_test.dart';
import 'package:phone/features/students/services/student_paid_lesson_calculator.dart';
import 'package:phone/generated/models/lesson_response.dart';
import 'package:phone/generated/models/student_payment_details_response.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/transaction_type.dart';
import 'package:phone/generated/models/visit_response.dart';
import 'package:phone/generated/models/visit_status.dart';
import 'package:phone/generated/models/visit_type.dart';
import 'package:phone/generated/models/visit_with_lesson_response.dart';

VisitWithLessonResponse _visit(
  DateTime date, {
  VisitStatus status = VisitStatus.present,
}) => VisitWithLessonResponse(
  visit: VisitResponse(
    id: 'v-${date.millisecondsSinceEpoch}',
    studentId: 'student-1',
    lessonId: 'l-${date.millisecondsSinceEpoch}',
    status: status,
    type: VisitType.regular,
    createdAt: date,
  ),
  lesson: LessonResponse(
    id: 'l-${date.millisecondsSinceEpoch}',
    date: date,
    createdAt: date,
  ),
);

TransactionResponse _subscription({
  required DateTime paidUntil,
  DateTime? paidFrom,
}) => TransactionResponse(
  id: 'tx-sub-${paidUntil.millisecondsSinceEpoch}',
  amount: 3000,
  date: paidFrom ?? paidUntil.subtract(const Duration(days: 30)),
  type: TransactionType.income,
  createdAt: paidFrom ?? paidUntil.subtract(const Duration(days: 30)),
  paymentDetails: StudentPaymentDetailsResponse(
    detailsType: 'STUDENT',
    studentId: 'student-1',
    paidFrom: paidFrom,
    paidUntil: paidUntil,
  ),
);

TransactionResponse _sessionsPayment(int count) => TransactionResponse(
  id: 'tx-ses-$count',
  amount: 4000,
  date: DateTime(2025, 8, 1),
  type: TransactionType.income,
  createdAt: DateTime(2025, 8, 1),
  paymentDetails: StudentPaymentDetailsResponse(
    detailsType: 'STUDENT',
    studentId: 'student-1',
    sessionsAdded: count,
  ),
);

void main() {
  group('findAutoPaidFromDate', () {
    test('Rule 1: if last paid until is bigger than current date, uses paid until + 1 day', () {
      final result = findAutoPaidFromDate(
        visits: [_visit(DateTime(2026, 11, 5))],
        payments: [_subscription(paidUntil: DateTime(2026, 11, 20))],
        currentDate: DateTime(2026, 11, 9),
      );

      expect(result, DateTime(2026, 11, 21));
    });

    test('Rule 1 applies even if there are no visits when paid until > current date', () {
      final result = findAutoPaidFromDate(
        visits: const [],
        payments: [_subscription(paidUntil: DateTime(2026, 11, 30))],
        currentDate: DateTime(2026, 11, 9),
      );

      expect(result, DateTime(2026, 12, 1));
    });

    test('Rule 2: if last paid lesson used session, uses last paid lesson date + 1 day', () {
      final result = findAutoPaidFromDate(
        visits: [_visit(DateTime(2025, 8, 10)), _visit(DateTime(2025, 8, 15))],
        payments: [_sessionsPayment(5)],
        currentDate: DateTime(2025, 8, 20),
      );

      // Last paid lesson is 2025-08-15, which used a session -> 2025-08-16
      expect(result, DateTime(2025, 8, 16));
    });

    test('Rule 2 ignores unpaid lessons when sessions ran out', () {
      final result = findAutoPaidFromDate(
        visits: [
          _visit(DateTime(2025, 8, 10)), // Paid via session 1
          _visit(DateTime(2025, 8, 15)), // Paid via session 2
          _visit(DateTime(2025, 8, 20)), // Unpaid (only 2 sessions available)
        ],
        payments: [_sessionsPayment(2)],
        currentDate: DateTime(2025, 8, 25),
      );

      // Last paid lesson is 2025-08-15 -> 2025-08-16
      expect(result, DateTime(2025, 8, 16));
    });

    test(
      'Rule 3: if last paid lesson used paid until, uses paid until + 1 day',
      () {
        final result = findAutoPaidFromDate(
          visits: [
            _visit(DateTime(2025, 8, 10)),
            _visit(DateTime(2025, 8, 25)),
          ],
          payments: [
            _subscription(
              paidFrom: DateTime(2025, 8, 1),
              paidUntil: DateTime(2025, 8, 31),
            ),
          ],
          currentDate: DateTime(2025, 9, 10),
        );

        // Last paid lesson is 2025-08-25, which used paidUntil (ending 2025-08-31) -> 2025-09-01
        expect(result, DateTime(2025, 9, 1));
      },
    );

    test('Rule 2 applies if a session was used after subscription expired', () {
      final result = findAutoPaidFromDate(
        visits: [
          _visit(DateTime(2025, 8, 25)), // Covered by subscription
          _visit(DateTime(2025, 9, 5)), // Covered by session
        ],
        payments: [
          _subscription(
            paidFrom: DateTime(2025, 8, 1),
            paidUntil: DateTime(2025, 8, 31),
          ),
          _sessionsPayment(2),
        ],
        currentDate: DateTime(2025, 9, 10),
      );

      // Last paid lesson was 2025-09-05 using session -> 2025-09-06
      expect(result, DateTime(2025, 9, 6));
    });

    test('returns null for free students', () {
      final result = findAutoPaidFromDate(
        visits: [_visit(DateTime(2025, 8, 10))],
        payments: [_sessionsPayment(5)],
        isStudentFree: true,
      );

      expect(result, isNull);
    });

    test('returns null when student has no payments or visits', () {
      final result = findAutoPaidFromDate(visits: const [], payments: const []);

      expect(result, isNull);
    });
  });
}
