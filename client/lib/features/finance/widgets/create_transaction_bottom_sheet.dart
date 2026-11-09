import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/core/errors/dio_error_mapper.dart';
import 'package:phone/core/widgets/app_snack_bar.dart';
import 'package:phone/core/widgets/date_input_field.dart';
import 'package:phone/features/finance/controllers/create_transaction_controller.dart';
import 'package:phone/features/finance/widgets/amount_field.dart';
import 'package:phone/features/finance/widgets/paid_from_field.dart';
import 'package:phone/features/finance/widgets/paid_until_field.dart';
import 'package:phone/features/finance/widgets/sessions_count_field.dart';
import 'package:phone/features/finance/widgets/sheet_title.dart';
import 'package:phone/features/finance/widgets/student_field.dart';
import 'package:phone/features/finance/widgets/student_payment_mode_selector.dart';
import 'package:phone/features/finance/widgets/submit_button.dart';
import 'package:phone/features/finance/widgets/transaction_type_selector.dart';
import 'package:phone/features/students/controllers/create_student_payment_controller.dart';
import 'package:phone/features/students/controllers/student_list_controller.dart';
import 'package:phone/features/students/repositories/student_repository.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_type.dart';
import 'package:phone/i18n/strings.g.dart';

class CreateTransactionBottomSheet extends ConsumerStatefulWidget {
  final String? initialStudentId;

  final String? initialStudentName;

  final bool isIncomeOnly;

  final bool isStudentLocked;

  final num? initialAmount;

  const new({
    super.key,
    this.initialStudentId,
    this.initialStudentName,
    this.isIncomeOnly = false,
    this.isStudentLocked = false,
    this.initialAmount,
  });

  @override
  ConsumerState<CreateTransactionBottomSheet> createState() =>
      _CreateTransactionBottomSheetState();
}

class _CreateTransactionBottomSheetState
    extends ConsumerState<CreateTransactionBottomSheet> {
  final TextEditingController _amountController = TextEditingController();
  final TextEditingController _sessionsController = TextEditingController();

  TransactionType _selectedType = TransactionType.income;
  StudentPaymentMode _paymentMode = StudentPaymentMode.paidUntil;
  DateTime _selectedDate = DateTime.now();
  bool _isDateValid = true;
  String? _selectedStudentId;
  DateTime? _paidFrom;
  DateTime? _paidUntil;
  bool _isLoading = false;
  bool _isLoadingPaidFrom = false;
  bool _isStudentExpanded = false;

  double? _autoFilledAmount;

  @override
  void initState() {
    super.initState();

    final initialAmount = widget.initialAmount?.toDouble();
    if (initialAmount != null) {
      _autoFilledAmount = initialAmount;
      _amountController.text = _formatAmount(initialAmount);
    }

    _amountController.addListener(_onAmountChanged);
    _sessionsController.addListener(_onSessionsChanged);

    final initialStudentId = widget.initialStudentId;
    if (initialStudentId != null) {
      _selectedStudentId = initialStudentId;
      _paidUntil = _defaultPaidUntil;
      _loadLastPaidLesson(initialStudentId);
    }
  }

  void _onAmountChanged() => setState(() {});

  void _onSessionsChanged() => setState(() {});

  @override
  void dispose() {
    _amountController.dispose();
    _sessionsController.dispose();
    super.dispose();
  }

  bool get _isIncome => _selectedType == TransactionType.income;

  bool get _isStudentPayment => _isIncome && _selectedStudentId != null;

  bool get _isPaidUntilMode => _paymentMode == StudentPaymentMode.paidUntil;

  int? get _parsedSessionsAdded =>
      int.tryParse(_sessionsController.text.trim());

  DateTime get _defaultPaidUntil =>
      DateTime(_selectedDate.year, _selectedDate.month + 1, _selectedDate.day);

  DateTime get _maxPaidUntil =>
      DateTime.now().add(const Duration(days: 365 * 3));

  void _onDateChanged(DateTime? date) {
    setState(() {
      _isDateValid = date != null;
      if (date == null) return;

      _selectedDate = date;
    });
  }

  DateTime _oneMonthAfter(DateTime date) =>
      DateTime(date.year, date.month + 1, date.day);

  void _onPaidUntilChanged(DateTime? date) {
    setState(() {
      _paidUntil = date;
    });
  }

  void _onPaidFromChanged(DateTime? date) {
    setState(() {
      _paidFrom = date;
      if (date != null) {
        _paidUntil = _oneMonthAfter(date);
      }
    });
  }

  Future<void> _loadLastPaidLesson(String studentId) async {
    setState(() {
      _isLoadingPaidFrom = true;
    });

    try {
      final lastPaid = await ref
          .read(studentRepositoryProvider)
          .getLastPaidLessonDate(studentId);
      if (!mounted || _selectedStudentId != studentId) return;
      setState(() {
        _paidFrom = lastPaid;
        if (lastPaid != null) {
          _paidUntil = _oneMonthAfter(lastPaid);
        }
      });
    } catch (_) {
    } finally {
      if (mounted && _selectedStudentId == studentId) {
        setState(() {
          _isLoadingPaidFrom = false;
        });
      }
    }
  }

  void _onPaymentModeChanged(StudentPaymentMode mode) {
    FocusScope.of(context).unfocus();
    setState(() {
      _paymentMode = mode;
      _isStudentExpanded = false;
      if (mode == StudentPaymentMode.paidUntil && _paidUntil == null) {
        _paidUntil = _paidFrom != null
            ? _oneMonthAfter(_paidFrom!)
            : _defaultPaidUntil;
      }
    });
  }

  void _onTypeChanged(TransactionType type) {
    setState(() {
      _selectedType = type;

      if (!_isIncome) {
        _isStudentExpanded = false;
      }
    });
  }

  void _toggleStudent() {
    FocusScope.of(context).unfocus();
    setState(() {
      _isStudentExpanded = !_isStudentExpanded;
    });
  }

  void _collapseStudent() {
    if (!_isStudentExpanded) return;

    setState(() {
      _isStudentExpanded = false;
    });
  }

  void _selectStudent(String? studentId) {
    final paidUntil = studentId == null
        ? null
        : (_paidUntil ?? _defaultPaidUntil);

    _applyGroupDefaultAmount(studentId);

    setState(() {
      _selectedStudentId = studentId;
      _isStudentExpanded = false;
      _paidUntil = paidUntil;
      _paidFrom = null;
    });

    if (studentId != null) {
      _loadLastPaidLesson(studentId);
    }
  }

  bool get _isDateRangeInvalid =>
      _paidFrom != null &&
      _paidUntil != null &&
      DateUtils.dateOnly(_paidFrom!).isAfter(DateUtils.dateOnly(_paidUntil!));

  Future<void> _onSave() async {
    final amountText = _amountController.text.trim();
    if (amountText.isEmpty) return;

    final amount = double.tryParse(amountText);
    if (amount == null || amount <= 0) return;

    final paidUntil = _isPaidUntilMode ? _paidUntil : null;
    final paidFrom = _isPaidUntilMode ? _paidFrom : null;
    final sessionsAdded = _isPaidUntilMode ? null : _parsedSessionsAdded;

    if (_isStudentPayment) {
      if (_isPaidUntilMode && (paidUntil == null || _isDateRangeInvalid)) {
        return;
      }
      if (!_isPaidUntilMode && (sessionsAdded == null || sessionsAdded <= 0)) {
        return;
      }
    }

    FocusScope.of(context).unfocus();

    setState(() {
      _isLoading = true;
      _isStudentExpanded = false;
    });

    try {
      final bool isSaved;

      if (_isStudentPayment) {
        isSaved = await ref
            .read(createStudentPaymentControllerProvider.notifier)
            .saveStudentPayment(
              studentId: _selectedStudentId!,
              amount: amount,
              date: _selectedDate,
              paidFrom: paidFrom,
              paidUntil: paidUntil,
              sessionsAdded: sessionsAdded,
            );
      } else {
        isSaved = await ref
            .read(createTransactionControllerProvider.notifier)
            .saveTransaction(
              type: _selectedType,
              amount: amount,
              date: _selectedDate,
            );
      }

      if (!mounted) return;

      if (!isSaved) {
        final error = _isStudentPayment
            ? ref.read(createStudentPaymentControllerProvider).error
            : ref.read(createTransactionControllerProvider).error;

        AppSnackBar.show(
          context,
          context.t.error,
          error?.userMessage(context.t) ?? context.t.error,
          SnackBarType.error,
        );
        return;
      }

      Navigator.of(context).pop();
    } catch (error) {
      if (!mounted) return;
      AppSnackBar.show(
        context,
        context.t.error,
        error.userMessage(context.t),
        SnackBarType.error,
      );
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  String _formatAmount(num amount) {
    final asDouble = amount.toDouble();

    return asDouble == asDouble.roundToDouble()
        ? asDouble.toInt().toString()
        : asDouble.toString();
  }

  double? _groupDefaultAmount(String? studentId) {
    if (studentId == null) return null;

    final students =
        ref.read(studentListControllerProvider).value ??
        const <StudentSummaryResponse>[];

    for (final student in students) {
      if (student.id == studentId) {
        return student.studentGroups?.firstOrNull?.defaultPrice?.toDouble();
      }
    }

    return null;
  }

  void _applyGroupDefaultAmount(String? studentId) {
    final defaultAmount = _groupDefaultAmount(studentId);
    if (defaultAmount == null || defaultAmount <= 0) return;

    final currentAmount = double.tryParse(_amountController.text.trim());
    final isAutoFilled =
        _autoFilledAmount != null && currentAmount == _autoFilledAmount;

    if (currentAmount != null && !isAutoFilled) return;

    _autoFilledAmount = defaultAmount;
    _amountController.text = _formatAmount(defaultAmount);
  }

  @override
  Widget build(BuildContext context) {
    final studentsState = ref.watch(studentListControllerProvider);
    final students = studentsState.value ?? const <StudentSummaryResponse>[];

    final amount = double.tryParse(_amountController.text.trim());
    final isAmountPositive = amount != null && amount > 0;
    final sessionsAdded = _parsedSessionsAdded;
    final isPaymentFieldInvalid =
        _isStudentPayment &&
        (_isPaidUntilMode
            ? (_paidUntil == null || _isDateRangeInvalid)
            : (sessionsAdded == null || sessionsAdded <= 0));
    final keyboardInset = MediaQuery.of(context).viewInsets.bottom;
    final canSubmit =
        isAmountPositive &&
        !isPaymentFieldInvalid &&
        _isDateValid &&
        !_isLoading;

    return SafeArea(
      child: Padding(
        padding: EdgeInsets.only(
          left: 22,
          right: 22,
          bottom: 8 + keyboardInset,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            SheetTitle(title: context.t.finance.createTransaction),
            const SizedBox(height: 16),
            Flexible(
              child: SingleChildScrollView(
                child: AbsorbPointer(
                  absorbing: _isLoading,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    spacing: 22,
                    children: [
                      if (!widget.isIncomeOnly)
                        TransactionTypeSelector(
                          selectedType: _selectedType,
                          onChanged: _onTypeChanged,
                        ),

                      AmountField(
                        controller: _amountController,
                        enabled: !_isLoading,
                        onTap: _collapseStudent,
                      ),

                      DateInputField(
                        labelText: '${context.t.finance.date}*',
                        icon: FluentIcons.calendar_32_regular,
                        initialDate: _selectedDate,
                        lastDate: DateTime.now(),
                        enabled: !_isLoading,
                        onDateChanged: _onDateChanged,
                      ),

                      if (_isIncome)
                        StudentField(
                          isLocked: widget.isStudentLocked,
                          lockedStudentName: widget.initialStudentName,
                          students: students,
                          isLoading: studentsState.isLoading,
                          selectedStudentId: _selectedStudentId,
                          isExpanded: _isStudentExpanded,
                          isSubmitting: _isLoading,
                          onToggle: _toggleStudent,
                          onSelect: _selectStudent,
                        ),

                      if (_isStudentPayment) ...[
                        StudentPaymentModeSelector(
                          selectedMode: _paymentMode,
                          onChanged: _onPaymentModeChanged,
                        ),
                        if (_isPaidUntilMode)
                          Row(
                            children: [
                              Expanded(
                                child: PaidFromField(
                                  initialDate: _paidFrom,
                                  lastDate: _maxPaidUntil,
                                  enabled: !_isLoading,
                                  isLoading: _isLoadingPaidFrom,
                                  onDateChanged: _onPaidFromChanged,
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: PaidUntilField(
                                  initialDate: _paidUntil,
                                  lastDate: _maxPaidUntil,
                                  enabled: !_isLoading,
                                  onDateChanged: _onPaidUntilChanged,
                                ),
                              ),
                            ],
                          )
                        else
                          SessionsCountField(
                            controller: _sessionsController,
                            enabled: !_isLoading,
                            onTap: _collapseStudent,
                          ),
                      ],

                      SubmitButton(
                        isLoading: _isLoading,
                        enabled: canSubmit,
                        onPressed: _onSave,
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
