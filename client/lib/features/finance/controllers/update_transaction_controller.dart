import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/core/extensions/date_time_extension.dart';
import 'package:phone/features/finance/controllers/transaction_mutation_refresh.dart';
import 'package:phone/features/finance/repositories/transaction_repository.dart';
import 'package:phone/generated/models/transaction_response.dart';

final updateTransactionControllerProvider =
    AsyncNotifierProvider<UpdateTransactionController, void>(
      UpdateTransactionController.new,
    );

class UpdateTransactionController extends AsyncNotifier<void> {
  @override
  FutureOr<void> build() {}

  Future<TransactionResponse?> updateTransaction({
    required String transactionId,
    required num amount,
    required DateTime date,
    String? studentId,
    DateTime? paidFrom,
    bool clearPaidFrom = false,
    DateTime? paidUntil,
    bool clearPaidUntil = false,
    int? sessionsAdded,
    bool clearSessionsAdded = false,
  }) async {
    TransactionResponse? updatedTransaction;

    state = const AsyncLoading();

    state = await AsyncValue.guard(() async {
      updatedTransaction = await ref
          .read(transactionRepositoryProvider)
          .updateTransaction(
            transactionId: transactionId,
            body: {
              'amount': amount,
              'date': date.toIso8601StringWithOffset(),
              if (paidFrom != null)
                'paidFrom': paidFrom.toIso8601StringWithOffset()
              else if (clearPaidFrom)
                'paidFrom': null,
              if (paidUntil != null)
                'paidUntil': paidUntil.toIso8601StringWithOffset()
              else if (clearPaidUntil)
                'paidUntil': null,
              if (sessionsAdded != null)
                'sessionsAdded': sessionsAdded
              else if (clearSessionsAdded)
                'sessionsAdded': null,
            },
          );

      await refreshAfterTransactionMutation(
        ref,
        action: 'updating',
        studentId: studentId,
      );
    });

    return state.hasError ? null : updatedTransaction;
  }
}
