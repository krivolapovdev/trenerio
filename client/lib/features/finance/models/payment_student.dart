import 'package:phone/generated/models/group_student_summary_response.dart';
import 'package:phone/generated/models/student_status.dart';
import 'package:phone/generated/models/student_summary_response.dart';

class PaymentStudent {
  final String id;
  final String fullName;
  final bool free;
  final List<StudentStatus> statuses;

  const new({
    required this.id,
    required this.fullName,
    required this.free,
    required this.statuses,
  });

  static PaymentStudent fromGroup(GroupStudentSummaryResponse student) =>
      PaymentStudent(
        id: student.id,
        fullName: student.fullName,
        free: student.free,
        statuses: student.statuses,
      );

  static PaymentStudent fromSummary(StudentSummaryResponse student) =>
      PaymentStudent(
        id: student.id,
        fullName: student.fullName,
        free: student.free,
        statuses: student.statuses,
      );
}
