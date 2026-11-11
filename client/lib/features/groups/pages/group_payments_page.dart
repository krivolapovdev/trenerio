import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/features/finance/models/payment_student.dart';
import 'package:phone/features/finance/widgets/payments/student_payments_view.dart';
import 'package:phone/features/groups/controllers/group_list_controller.dart';
import 'package:phone/features/groups/controllers/group_payments_controller.dart';
import 'package:phone/features/groups/controllers/group_students_controller.dart';
import 'package:phone/generated/models/group_student_summary_response.dart';
import 'package:phone/generated/models/group_summary_response.dart';
import 'package:phone/i18n/strings.g.dart';

class GroupPaymentsPage extends ConsumerWidget {
  final GroupSummaryResponse group;

  const new({super.key, required this.group});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final groupName =
        ref
            .watch(groupListControllerProvider)
            .value
            ?.where((g) => g.id == group.id)
            .firstOrNull
            ?.name ??
        group.name;

    final studentsState = ref.watch(groupStudentsProvider(group.id));
    final isLoading = ref.watch(groupPaymentsControllerProvider).isLoading;
    final rawStudents =
        studentsState.value ?? const <GroupStudentSummaryResponse>[];
    final students = rawStudents.map(PaymentStudent.fromGroup).toList();

    return StudentPaymentsView(
      title: groupName.isNotEmpty
          ? context.t.groups.groupPaymentsTitle(name: groupName)
          : context.t.groups.payments,
      students: students,
      isLoading: isLoading,
      isStudentsLoading: studentsState.isLoading,
      defaultPrice: group.defaultPrice,
      onSave: ({required paymentDate, required payments}) => ref
          .read(groupPaymentsControllerProvider.notifier)
          .createPayments(
            groupId: group.id,
            paymentDate: paymentDate,
            payments: payments,
          ),
      error: ref.watch(groupPaymentsControllerProvider).error,
    );
  }
}
