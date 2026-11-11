import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/features/finance/models/payment_student.dart';
import 'package:phone/features/finance/widgets/payments/student_payments_view.dart';
import 'package:phone/features/students/controllers/student_list_controller.dart';
import 'package:phone/features/students/controllers/student_payments_batch_controller.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/i18n/strings.g.dart';

class StudentPaymentsPage extends ConsumerWidget {
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final studentsState = ref.watch(studentListControllerProvider);
    final isLoading = ref
        .watch(studentPaymentsBatchControllerProvider)
        .isLoading;
    final rawStudents = studentsState.value ?? const <StudentSummaryResponse>[];
    final students = rawStudents.map(PaymentStudent.fromSummary).toList();

    return StudentPaymentsView(
      title: context.t.students.payments,
      students: students,
      isLoading: isLoading,
      isStudentsLoading: studentsState.isLoading,
      onSave: ({required paymentDate, required payments}) => ref
          .read(studentPaymentsBatchControllerProvider.notifier)
          .createPayments(paymentDate: paymentDate, payments: payments),
      error: ref.watch(studentPaymentsBatchControllerProvider).error,
    );
  }
}
