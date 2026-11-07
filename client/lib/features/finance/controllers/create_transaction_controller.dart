import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/features/finance/controllers/transaction_mutation_refresh.dart';
import 'package:phone/features/finance/repositories/transaction_repository.dart';
import 'package:phone/generated/models/create_student_payment_details_request.dart';
import 'package:phone/generated/models/transaction_type.dart';

final createTransactionControllerProvider =
    AsyncNotifierProvider<CreateTransactionController, void>(
      CreateTransactionController.new,
    );

class CreateTransactionController extends AsyncNotifier<void> {
  @override
  Future<void> build() async {}

  Future<bool> saveTransaction({
    required TransactionType type,
    required double amount,
    required DateTime date,
    CreateStudentPaymentDetailsRequest? paymentDetails,
  }) async {
    final transactionRepository = ref.read(transactionRepositoryProvider);

    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      await transactionRepository.createTransaction(
        type: type,
        amount: amount,
        date: date,
        paymentDetails: paymentDetails,
      );

      await refreshAfterTransactionMutation(
        ref,
        action: 'creating',
        studentId: paymentDetails?.studentId,
      );
    });

    return !state.hasError;
  }
}
