import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/features/finance/models/student_payment.dart';
import 'package:phone/features/groups/controllers/group_mutation_refresh.dart';
import 'package:phone/features/students/repositories/student_repository.dart';

final studentPaymentsBatchControllerProvider =
    AsyncNotifierProvider<StudentPaymentsBatchController, void>(
      StudentPaymentsBatchController.new,
    );

class StudentPaymentsBatchController extends AsyncNotifier<void> {
  @override
  FutureOr<void> build() {}

  Future<bool> createPayments({
    required DateTime paymentDate,
    required List<StudentPaymentItem> payments,
  }) async {
    state = const AsyncLoading();

    state = await AsyncValue.guard(() async {
      final studentRepo = ref.read(studentRepositoryProvider);

      for (final payment in payments) {
        await studentRepo.createStudentPayment(
          studentId: payment.studentId,
          amount: payment.amount,
          date: paymentDate,
          paidFrom: payment.paidFrom,
          paidUntil: payment.paidUntil,
          sessionsAdded: payment.sessionsAdded,
        );
      }

      await refreshAfterStudentPaymentsMutation(
        ref,
        studentIds: payments.map((p) => p.studentId),
      );
    });

    return !state.hasError;
  }
}
