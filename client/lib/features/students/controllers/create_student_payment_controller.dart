import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/features/finance/controllers/transaction_mutation_refresh.dart';
import 'package:phone/features/students/repositories/student_repository.dart';

final createStudentPaymentControllerProvider =
    AsyncNotifierProvider<CreateStudentPaymentController, void>(
      CreateStudentPaymentController.new,
    );

class CreateStudentPaymentController extends AsyncNotifier<void> {
  @override
  FutureOr<void> build() {}

  Future<bool> saveStudentPayment({
    required String studentId,
    required double amount,
    required DateTime date,
    DateTime? paidUntil,
    int? sessionsAdded,
  }) async {
    state = const AsyncLoading();

    state = await AsyncValue.guard(() async {
      if (paidUntil == null && (sessionsAdded == null || sessionsAdded <= 0)) {
        throw ArgumentError(
          'Either paidUntil or a positive sessionsAdded must be provided.',
        );
      }

      await ref
          .read(studentRepositoryProvider)
          .createStudentPayment(
            studentId: studentId,
            amount: amount,
            date: date,
            paidUntil: paidUntil,
            sessionsAdded: sessionsAdded,
          );

      await refreshAfterTransactionMutation(
        ref,
        action: 'creating',
        studentId: studentId,
      );
    });

    return !state.hasError;
  }
}
