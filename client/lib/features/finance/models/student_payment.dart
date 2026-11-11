import 'package:phone/features/finance/widgets/student_payment_mode_selector.dart';

class StudentPaymentRecord {
  final bool selected;
  final StudentPaymentMode? customPaymentMode;
  final double? customAmount;
  final DateTime? customPaidFrom;
  final DateTime? customPaidUntil;
  final int? customSessionsAdded;

  const new({
    this.selected = true,
    this.customPaymentMode,
    this.customAmount,
    this.customPaidFrom,
    this.customPaidUntil,
    this.customSessionsAdded,
  });

  bool get hasCustomValues =>
      customPaymentMode != null ||
      customAmount != null ||
      customPaidFrom != null ||
      customPaidUntil != null ||
      customSessionsAdded != null;

  StudentPaymentRecord copyWith({
    bool? selected,
    StudentPaymentMode? customPaymentMode,
    bool clearCustomPaymentMode = false,
    double? customAmount,
    bool clearCustomAmount = false,
    DateTime? customPaidFrom,
    bool clearCustomPaidFrom = false,
    DateTime? customPaidUntil,
    bool clearCustomPaidUntil = false,
    int? customSessionsAdded,
    bool clearCustomSessionsAdded = false,
  }) => StudentPaymentRecord(
    selected: selected ?? this.selected,
    customPaymentMode: clearCustomPaymentMode
        ? null
        : (customPaymentMode ?? this.customPaymentMode),
    customAmount: clearCustomAmount
        ? null
        : (customAmount ?? this.customAmount),
    customPaidFrom: clearCustomPaidFrom
        ? null
        : (customPaidFrom ?? this.customPaidFrom),
    customPaidUntil: clearCustomPaidUntil
        ? null
        : (customPaidUntil ?? this.customPaidUntil),
    customSessionsAdded: clearCustomSessionsAdded
        ? null
        : (customSessionsAdded ?? this.customSessionsAdded),
  );
}

class StudentPaymentItem {
  final String studentId;
  final double amount;
  final DateTime? paidFrom;
  final DateTime? paidUntil;
  final int? sessionsAdded;

  const new({
    required this.studentId,
    required this.amount,
    this.paidFrom,
    this.paidUntil,
    this.sessionsAdded,
  });
}

typedef GroupStudentPayment = StudentPaymentRecord;
typedef GroupStudentPaymentItem = StudentPaymentItem;
