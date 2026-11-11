import 'package:flutter/material.dart';
import 'package:phone/core/widgets/app_bottom_sheet.dart';
import 'package:phone/core/widgets/app_snack_bar.dart';
import 'package:phone/features/finance/models/payment_student.dart';
import 'package:phone/features/finance/models/student_payment.dart';
import 'package:phone/features/finance/widgets/payments/payment_amount_card.dart';
import 'package:phone/features/finance/widgets/payments/payment_date_card.dart';
import 'package:phone/features/finance/widgets/payments/payment_sessions_card.dart';
import 'package:phone/features/finance/widgets/payments/save_payments_action.dart';
import 'package:phone/features/finance/widgets/payments/student_payment_bottom_sheet.dart';
import 'package:phone/features/finance/widgets/payments/student_payment_card.dart';
import 'package:phone/features/finance/widgets/student_payment_mode_selector.dart';
import 'package:phone/generated/models/student_status.dart';
import 'package:phone/i18n/strings.g.dart';

class StudentPaymentsView extends StatefulWidget {
  final String title;
  final List<PaymentStudent> students;
  final bool isLoading;
  final bool isStudentsLoading;
  final num? defaultPrice;
  final Future<bool> Function({
    required DateTime paymentDate,
    required List<StudentPaymentItem> payments,
  })
  onSave;
  final Object? error;

  const new({
    super.key,
    required this.title,
    required this.students,
    required this.isLoading,
    this.isStudentsLoading = false,
    this.defaultPrice,
    required this.onSave,
    this.error,
  });

  @override
  State<StudentPaymentsView> createState() => _StudentPaymentsViewState();
}

class _StudentPaymentsViewState extends State<StudentPaymentsView> {
  late DateTime _paymentDate;
  DateTime? _paidFromDate;
  late DateTime _paidUntilDate;
  late final TextEditingController _amountController;
  late final TextEditingController _sessionsController;
  StudentPaymentMode _paymentMode = StudentPaymentMode.paidUntil;

  Map<String, StudentPaymentRecord>? _payments;

  @override
  void initState() {
    super.initState();
    _paymentDate = DateTime.now();
    _paidFromDate = DateUtils.dateOnly(_paymentDate);
    _paidUntilDate = DateTime(
      _paymentDate.year,
      _paymentDate.month + 1,
      _paymentDate.day,
    );

    final defaultPrice = widget.defaultPrice;
    final initialAmount = (defaultPrice != null && defaultPrice > 0)
        ? _formatAmount(defaultPrice)
        : '';
    _amountController = TextEditingController(text: initialAmount);
    _amountController.addListener(_onFieldChanged);

    _sessionsController = TextEditingController();
    _sessionsController.addListener(_onFieldChanged);
  }

  @override
  void dispose() {
    _amountController.removeListener(_onFieldChanged);
    _amountController.dispose();
    _sessionsController.removeListener(_onFieldChanged);
    _sessionsController.dispose();
    super.dispose();
  }

  void _onFieldChanged() => setState(() {});

  String _formatAmount(num amount) {
    final asDouble = amount.toDouble();
    return asDouble == asDouble.roundToDouble()
        ? asDouble.toInt().toString()
        : asDouble.toString();
  }

  String _formatDate(DateTime date) {
    final day = date.day.toString().padLeft(2, '0');
    final month = date.month.toString().padLeft(2, '0');
    return '$day.$month.${date.year}';
  }

  bool get _isDateRangeInvalid =>
      _paymentMode == StudentPaymentMode.paidUntil &&
      _paidFromDate != null &&
      DateUtils.dateOnly(_paidFromDate!)
          .isAfter(DateUtils.dateOnly(_paidUntilDate));

  int? get _parsedDefaultSessions =>
      int.tryParse(_sessionsController.text.trim());

  Map<String, StudentPaymentRecord> _effectivePayments(
    List<PaymentStudent> students,
  ) {
    final current = _payments;
    if (current != null) return current;

    final hasAnyUnpaid = students.any(
      (s) => !s.free && !s.statuses.contains(StudentStatus.paid),
    );

    final initial = {
      for (final student in students)
        student.id: StudentPaymentRecord(
          selected:
              !student.free &&
              (hasAnyUnpaid
                  ? !student.statuses.contains(StudentStatus.paid)
                  : true),
        ),
    };
    if (students.isNotEmpty) {
      _payments = initial;
    }
    return initial;
  }

  Future<void> _pickPaymentDate() async {
    final today = DateUtils.dateOnly(DateTime.now());
    final selected = DateUtils.dateOnly(_paymentDate);
    final picked = await showDatePicker(
      context: context,
      initialDate: selected.isAfter(today) ? today : selected,
      firstDate: DateTime(2000),
      lastDate: today,
    );

    if (picked == null || !mounted) return;

    setState(() {
      _paymentDate = picked;
    });
  }

  Future<void> _pickPaidFromDate() async {
    final firstDate = DateTime(2000);
    final lastDate = DateUtils.dateOnly(
      DateTime.now().add(const Duration(days: 365 * 3)),
    );
    final selected = DateUtils.dateOnly(_paidFromDate ?? _paymentDate);
    final initial = selected.isBefore(firstDate)
        ? firstDate
        : (selected.isAfter(lastDate) ? lastDate : selected);

    final picked = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: firstDate,
      lastDate: lastDate,
    );

    if (picked == null || !mounted) return;

    setState(() {
      _paidFromDate = picked;
    });
  }

  Future<void> _pickPaidUntilDate() async {
    final firstDate = DateTime(2000);
    final lastDate = DateUtils.dateOnly(
      DateTime.now().add(const Duration(days: 365 * 3)),
    );
    final selected = DateUtils.dateOnly(_paidUntilDate);
    final initial = selected.isBefore(firstDate)
        ? firstDate
        : (selected.isAfter(lastDate) ? lastDate : selected);

    final picked = await showDatePicker(
      context: context,
      initialDate: initial,
      firstDate: firstDate,
      lastDate: lastDate,
    );

    if (picked == null || !mounted) return;

    setState(() {
      _paidUntilDate = picked;
    });
  }

  void _toggleStudent(
    Map<String, StudentPaymentRecord> payments,
    PaymentStudent student,
  ) {
    setState(() {
      final current = payments[student.id];
      final isSelected = current?.selected ?? false;
      _payments = {
        ...payments,
        student.id: (current ?? const StudentPaymentRecord()).copyWith(
          selected: !isSelected,
        ),
      };
    });
  }

  void _toggleAll(
    List<PaymentStudent> students,
    Map<String, StudentPaymentRecord> payments,
  ) {
    setState(() {
      final payingStudents = students.where((s) => !s.free).toList();
      final allSelected =
          payingStudents.isNotEmpty &&
          payingStudents.every((s) => payments[s.id]?.selected ?? false);

      final next = {...payments};
      for (final student in payingStudents) {
        final current = next[student.id] ?? const StudentPaymentRecord();
        next[student.id] = current.copyWith(selected: !allSelected);
      }
      _payments = next;
    });
  }

  Future<void> _editStudent(
    PaymentStudent student,
    Map<String, StudentPaymentRecord> payments,
  ) async {
    final currentPayment = payments[student.id];
    final defaultAmount = double.tryParse(_amountController.text.trim()) ?? 0;

    final updated = await AppBottomSheet.show<StudentPaymentRecord>(
      context: context,
      child: StudentPaymentBottomSheet(
        studentName: student.fullName,
        defaultPaymentMode: _paymentMode,
        defaultAmount: defaultAmount,
        defaultPaidFrom: _paidFromDate,
        defaultPaidUntil: _paidUntilDate,
        defaultSessions: _parsedDefaultSessions,
        currentCustomMode: currentPayment?.customPaymentMode,
        currentCustomAmount: currentPayment?.customAmount,
        currentCustomPaidFrom: currentPayment?.customPaidFrom,
        currentCustomPaidUntil: currentPayment?.customPaidUntil,
        currentCustomSessions: currentPayment?.customSessionsAdded,
      ),
    );

    if (!mounted || updated == null) return;

    setState(() {
      _payments = {
        ...payments,
        student.id: (currentPayment ?? const StudentPaymentRecord()).copyWith(
          selected: true,
          customPaymentMode: updated.customPaymentMode,
          clearCustomPaymentMode: updated.customPaymentMode == null,
          customAmount: updated.customAmount,
          clearCustomAmount: updated.customAmount == null,
          customPaidFrom: updated.customPaidFrom,
          clearCustomPaidFrom: updated.customPaidFrom == null,
          customPaidUntil: updated.customPaidUntil,
          clearCustomPaidUntil: updated.customPaidUntil == null,
          customSessionsAdded: updated.customSessionsAdded,
          clearCustomSessionsAdded: updated.customSessionsAdded == null,
        ),
      };
    });
  }

  Future<void> _onSubmit(
    List<PaymentStudent> students,
    Map<String, StudentPaymentRecord> payments,
  ) async {
    FocusScope.of(context).unfocus();

    final defaultAmount = double.tryParse(_amountController.text.trim());
    final defaultSessions = _parsedDefaultSessions;
    final items = <StudentPaymentItem>[];

    for (final student in students) {
      final payment = payments[student.id];
      if (payment == null || !payment.selected) continue;

      final amount = payment.customAmount ?? defaultAmount;
      if (amount == null || amount <= 0) continue;

      final mode = payment.customPaymentMode ?? _paymentMode;
      if (mode == StudentPaymentMode.paidUntil) {
        final paidUntil = payment.customPaidUntil ?? _paidUntilDate;
        final paidFrom = payment.customPaidFrom ?? _paidFromDate;
        if (paidFrom != null &&
            DateUtils.dateOnly(paidFrom)
                .isAfter(DateUtils.dateOnly(paidUntil))) {
          continue;
        }

        items.add(
          StudentPaymentItem(
            studentId: student.id,
            amount: amount,
            paidFrom: paidFrom,
            paidUntil: paidUntil,
          ),
        );
      } else {
        final sessions = payment.customSessionsAdded ?? defaultSessions;
        if (sessions == null || sessions <= 0) continue;

        items.add(
          StudentPaymentItem(
            studentId: student.id,
            amount: amount,
            sessionsAdded: sessions,
          ),
        );
      }
    }

    if (items.isEmpty) return;

    final isSaved = await widget.onSave(
      paymentDate: _paymentDate,
      payments: items,
    );

    if (!mounted) return;

    if (isSaved) {
      Navigator.of(context).pop();
      return;
    }

    final error = widget.error;

    AppSnackBar.show(
      context,
      context.t.error,
      error?.toString() ?? context.t.error,
      SnackBarType.error,
    );
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final payments = _effectivePayments(widget.students);
    final defaultAmount = double.tryParse(_amountController.text.trim());

    final selectedStudents = widget.students
        .where((s) => payments[s.id]?.selected ?? false)
        .toList();
    final selectedCount = selectedStudents.length;

    final canSubmit =
        !widget.isLoading &&
        selectedCount > 0 &&
        !_isDateRangeInvalid &&
        selectedStudents.every((s) {
          final p = payments[s.id];
          final amount = p?.customAmount ?? defaultAmount;
          if (amount == null || amount <= 0) return false;

          final mode = p?.customPaymentMode ?? _paymentMode;
          if (mode == StudentPaymentMode.paidUntil) {
            final paidUntil = p?.customPaidUntil ?? _paidUntilDate;
            final paidFrom = p?.customPaidFrom ?? _paidFromDate;
            if (paidFrom != null &&
                DateUtils.dateOnly(paidFrom)
                    .isAfter(DateUtils.dateOnly(paidUntil))) {
              return false;
            }
            return true;
          } else {
            final sessions = p?.customSessionsAdded ?? _parsedDefaultSessions;
            return sessions != null && sessions > 0;
          }
        });

    final canEdit = !widget.isLoading && !widget.isStudentsLoading;

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        actions: [
          SavePaymentsAction(
            isLoading: widget.isLoading,
            enabled: canSubmit,
            onPressed: () => _onSubmit(widget.students, payments),
          ),
          const SizedBox(width: 8),
        ],
        backgroundColor: colorScheme.surface,
        surfaceTintColor: colorScheme.surface,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(bottom: Radius.circular(16)),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(22, 16, 22, 16),
        child: AbsorbPointer(
          absorbing: widget.isLoading,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            spacing: 16,
            children: [
              PaymentDateCard(
                label: context.t.finance.date,
                formattedDate: _formatDate(_paymentDate),
                canEdit: canEdit,
                onTap: _pickPaymentDate,
              ),

              Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                ),
                padding: const EdgeInsets.all(8.0),
                child: StudentPaymentModeSelector(
                  selectedMode: _paymentMode,
                  onChanged: canEdit
                      ? (mode) => setState(() => _paymentMode = mode)
                      : (_) {},
                ),
              ),

              if (_paymentMode == StudentPaymentMode.paidUntil)
                Row(
                  children: [
                    Expanded(
                      child: PaymentDateCard(
                        label: context.t.finance.paidFrom,
                        formattedDate: _paidFromDate != null
                            ? _formatDate(_paidFromDate!)
                            : '',
                        canEdit: canEdit,
                        allowClear: true,
                        errorText: _isDateRangeInvalid
                            ? context.t.invalidDate
                            : null,
                        onTap: _pickPaidFromDate,
                        onClear: () => setState(() => _paidFromDate = null),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: PaymentDateCard(
                        label: '${context.t.finance.paidUntil}*',
                        formattedDate: _formatDate(_paidUntilDate),
                        canEdit: canEdit,
                        errorText: _isDateRangeInvalid
                            ? context.t.invalidDate
                            : null,
                        onTap: _pickPaidUntilDate,
                      ),
                    ),
                  ],
                )
              else
                PaymentSessionsCard(
                  controller: _sessionsController,
                  canEdit: canEdit,
                ),
              PaymentAmountCard(
                controller: _amountController,
                canEdit: canEdit,
              ),
              StudentPaymentCard(
                students: widget.students,
                payments: payments,
                canEdit: canEdit,
                isLoading: widget.isStudentsLoading,
                onToggle: (student) => _toggleStudent(payments, student),
                onEdit: (student) => _editStudent(student, payments),
                onToggleAll: () => _toggleAll(widget.students, payments),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
