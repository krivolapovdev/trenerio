import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/core/errors/dio_error_mapper.dart';
import 'package:phone/core/widgets/app_snack_bar.dart';
import 'package:phone/core/widgets/date_input_field.dart';
import 'package:phone/features/finance/controllers/update_transaction_controller.dart';
import 'package:phone/features/finance/widgets/paid_from_field.dart';
import 'package:phone/features/finance/widgets/paid_until_field.dart';
import 'package:phone/features/finance/widgets/sessions_count_field.dart';
import 'package:phone/features/finance/widgets/student_payment_mode_selector.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/transaction_type.dart';
import 'package:phone/i18n/strings.g.dart';

class EditTransactionBottomSheet extends ConsumerStatefulWidget {
  final TransactionResponse transaction;

  const new({super.key, required this.transaction});

  @override
  ConsumerState<EditTransactionBottomSheet> createState() =>
      _EditTransactionBottomSheetState();
}

class _EditTransactionBottomSheetState
    extends ConsumerState<EditTransactionBottomSheet> {
  final TextEditingController _amountController = TextEditingController();
  final TextEditingController _sessionsController = TextEditingController();

  late DateTime _selectedDate;
  late DateTime? _paidUntil;
  DateTime? _paidFrom;
  late StudentPaymentMode _paymentMode;

  late final String _initialAmountText;
  late final StudentPaymentMode _initialPaymentMode;

  bool _isDateValid = true;
  bool _isLoading = false;

  bool get _isStudentPayment => widget.transaction.paymentDetails != null;

  bool get _isPaidUntilMode => _paymentMode == StudentPaymentMode.paidUntil;

  int? get _parsedSessionsAdded =>
      int.tryParse(_sessionsController.text.trim());

  DateTime get _defaultPaidUntil =>
      DateTime(_selectedDate.year, _selectedDate.month + 1, _selectedDate.day);

  DateTime get _maxPaidUntil =>
      DateTime.now().add(const Duration(days: 365 * 3));

  bool get _hasChanges =>
      _amountController.text != _initialAmountText ||
      DateUtils.dateOnly(_selectedDate) !=
          DateUtils.dateOnly(widget.transaction.date) ||
      _hasPaymentDetailsChanged;

  bool get _hasPaymentDetailsChanged {
    if (!_isStudentPayment) return false;
    if (_paymentMode != _initialPaymentMode) return true;

    if (_isPaidUntilMode) {
      final initialPaidUntil = widget.transaction.paymentDetails?.paidUntil;
      final paidUntil = _paidUntil;

      final isPaidUntilChanged = (initialPaidUntil == null || paidUntil == null)
          ? initialPaidUntil != paidUntil
          : DateUtils.dateOnly(paidUntil) !=
                DateUtils.dateOnly(initialPaidUntil);

      final initialPaidFrom = widget.transaction.paymentDetails?.paidFrom;
      final paidFrom = _paidFrom;

      final isPaidFromChanged = (initialPaidFrom == null || paidFrom == null)
          ? initialPaidFrom != paidFrom
          : DateUtils.dateOnly(paidFrom) != DateUtils.dateOnly(initialPaidFrom);

      return isPaidUntilChanged || isPaidFromChanged;
    }

    return _parsedSessionsAdded !=
        widget.transaction.paymentDetails?.sessionsAdded;
  }

  bool get _isDateRangeInvalid =>
      _paidFrom != null &&
      _paidUntil != null &&
      DateUtils.dateOnly(_paidFrom!).isAfter(DateUtils.dateOnly(_paidUntil!));

  bool get _isPaymentFieldInvalid {
    if (!_isStudentPayment) return false;
    if (_isPaidUntilMode) return _paidUntil == null || _isDateRangeInvalid;
    final sessions = _parsedSessionsAdded;
    return sessions == null || sessions <= 0;
  }

  @override
  void initState() {
    super.initState();

    final amount = widget.transaction.amount;
    _initialAmountText = amount % 1 == 0
        ? amount.toInt().toString()
        : amount.toString();

    _amountController
      ..text = _initialAmountText
      ..addListener(_onAmountChanged);

    final details = widget.transaction.paymentDetails;
    _selectedDate = widget.transaction.date;
    _paidUntil = details?.paidUntil;
    _paidFrom = details?.paidFrom;

    final initialSessions = details?.sessionsAdded;
    final initialSessionsText = initialSessions != null && initialSessions > 0
        ? initialSessions.toString()
        : '';

    _initialPaymentMode = (_paidUntil == null && initialSessions != null)
        ? StudentPaymentMode.sessionsCount
        : StudentPaymentMode.paidUntil;
    _paymentMode = _initialPaymentMode;

    _sessionsController
      ..text = initialSessionsText
      ..addListener(_onSessionsChanged);
  }

  void _onAmountChanged() {
    setState(() {});
  }

  void _onSessionsChanged() {
    setState(() {});
  }

  @override
  void dispose() {
    _amountController.dispose();
    _sessionsController.dispose();
    super.dispose();
  }

  void _onDateChanged(DateTime? date) {
    setState(() {
      _isDateValid = date != null;
      if (date == null) return;

      _selectedDate = date;
    });
  }

  void _onPaidUntilChanged(DateTime? date) {
    setState(() {
      _paidUntil = date;
    });
  }

  DateTime _oneMonthAfter(DateTime date) =>
      DateTime(date.year, date.month + 1, date.day);

  void _onPaidFromChanged(DateTime? date) {
    setState(() {
      _paidFrom = date;
      if (date != null) {
        _paidUntil = _oneMonthAfter(date);
      }
    });
  }

  void _onPaymentModeChanged(StudentPaymentMode mode) {
    FocusScope.of(context).unfocus();
    setState(() {
      _paymentMode = mode;
      if (mode == StudentPaymentMode.paidUntil) {
        _paidUntil ??=
            widget.transaction.paymentDetails?.paidUntil ?? _defaultPaidUntil;
        _paidFrom ??= widget.transaction.paymentDetails?.paidFrom;
      }
    });
  }

  Future<void> _onUpdate() async {
    final amount = double.tryParse(_amountController.text.trim());
    if (amount == null || amount <= 0) return;
    if (_isPaymentFieldInvalid) return;

    final details = widget.transaction.paymentDetails;
    final paidFrom = _isStudentPayment && _isPaidUntilMode ? _paidFrom : null;
    final clearPaidFrom =
        _isStudentPayment &&
        ((!_isPaidUntilMode && details?.paidFrom != null) ||
            (_isPaidUntilMode &&
                _paidFrom == null &&
                details?.paidFrom != null));
    final paidUntil = _isStudentPayment && _isPaidUntilMode ? _paidUntil : null;
    final clearPaidUntil =
        _isStudentPayment && !_isPaidUntilMode && details?.paidUntil != null;
    final sessionsAdded = _isStudentPayment && !_isPaidUntilMode
        ? _parsedSessionsAdded
        : null;
    final clearSessionsAdded =
        _isStudentPayment && _isPaidUntilMode && details?.sessionsAdded != null;

    FocusScope.of(context).unfocus();

    setState(() {
      _isLoading = true;
    });

    final updated = await ref
        .read(updateTransactionControllerProvider.notifier)
        .updateTransaction(
          transactionId: widget.transaction.id,
          amount: amount,
          date: _selectedDate,
          studentId: details?.studentId,
          paidFrom: paidFrom,
          clearPaidFrom: clearPaidFrom,
          paidUntil: paidUntil,
          clearPaidUntil: clearPaidUntil,
          sessionsAdded: sessionsAdded,
          clearSessionsAdded: clearSessionsAdded,
        );

    if (!mounted) return;

    if (updated != null) {
      Navigator.of(context).pop(updated);
      return;
    }

    setState(() {
      _isLoading = false;
    });

    final error = ref.read(updateTransactionControllerProvider).error;
    if (error != null) {
      AppSnackBar.show(
        context,
        context.t.error,
        error.userMessage(context.t),
        SnackBarType.error,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final parsedAmount = double.tryParse(_amountController.text.trim());
    final isAmountInvalid = parsedAmount == null || parsedAmount <= 0;
    final isIncome = widget.transaction.type == TransactionType.income;
    final keyboardInset = MediaQuery.of(context).viewInsets.bottom;

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
            Text(
              context.t.finance.editTransaction,
              style: const TextStyle(fontSize: 20, color: Colors.black),
              textAlign: TextAlign.center,
            ),

            const SizedBox(height: 8),

            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  isIncome
                      ? FluentIcons.arrow_down_24_regular
                      : FluentIcons.arrow_up_24_regular,
                  color: isIncome ? Colors.green : Colors.redAccent,
                ),
                const SizedBox(width: 8),
                Text(
                  isIncome
                      ? context.t.finance.typeIncome
                      : context.t.finance.typeExpense,
                  style: const TextStyle(fontSize: 16, color: Colors.black87),
                ),
              ],
            ),

            const SizedBox(height: 16),

            Flexible(
              child: SingleChildScrollView(
                child: AbsorbPointer(
                  absorbing: _isLoading,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    spacing: 22,
                    children: [
                      TextField(
                        controller: _amountController,
                        enabled: !_isLoading,
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        inputFormatters: [
                          FilteringTextInputFormatter.allow(
                            RegExp(r'^\d*\.?\d{0,2}'),
                          ),
                        ],
                        decoration: InputDecoration(
                          labelText: context.t.finance.amount,
                          hintText: '0.00',
                          prefixIcon: const Icon(FluentIcons.money_24_regular),
                        ),
                      ),

                      DateInputField(
                        labelText: context.t.finance.date,
                        icon: FluentIcons.calendar_32_regular,
                        initialDate: _selectedDate,
                        lastDate: DateTime.now(),
                        enabled: !_isLoading,
                        onDateChanged: _onDateChanged,
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
                          ),
                      ],

                      Padding(
                        padding: const EdgeInsets.symmetric(vertical: 10),
                        child: Align(
                          alignment: Alignment.centerRight,
                          child: TextButton.icon(
                            onPressed:
                                (isAmountInvalid ||
                                    !_isDateValid ||
                                    _isPaymentFieldInvalid ||
                                    !_hasChanges ||
                                    _isLoading)
                                ? null
                                : _onUpdate,
                            icon: _isLoading
                                ? const SizedBox(
                                    width: 18,
                                    height: 18,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  )
                                : null,
                            label: Text(
                              context.t.update,
                              style: const TextStyle(fontSize: 18),
                            ),
                          ),
                        ),
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
