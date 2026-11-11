import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:phone/features/finance/widgets/paid_from_field.dart';
import 'package:phone/features/finance/widgets/paid_until_field.dart';
import 'package:phone/features/finance/widgets/sessions_count_field.dart';
import 'package:phone/features/finance/widgets/student_payment_mode_selector.dart';
import 'package:phone/features/groups/models/group_student_payment.dart';
import 'package:phone/i18n/strings.g.dart';

class StudentPaymentBottomSheet extends StatefulWidget {
  final String studentName;
  final StudentPaymentMode defaultPaymentMode;
  final double defaultAmount;
  final DateTime? defaultPaidFrom;
  final DateTime defaultPaidUntil;
  final int? defaultSessions;
  final StudentPaymentMode? currentCustomMode;
  final double? currentCustomAmount;
  final DateTime? currentCustomPaidFrom;
  final DateTime? currentCustomPaidUntil;
  final int? currentCustomSessions;

  const new({
    super.key,
    required this.studentName,
    required this.defaultPaymentMode,
    required this.defaultAmount,
    this.defaultPaidFrom,
    required this.defaultPaidUntil,
    this.defaultSessions,
    this.currentCustomMode,
    this.currentCustomAmount,
    this.currentCustomPaidFrom,
    this.currentCustomPaidUntil,
    this.currentCustomSessions,
  });

  @override
  State<StudentPaymentBottomSheet> createState() =>
      _StudentPaymentBottomSheetState();
}

class _StudentPaymentBottomSheetState extends State<StudentPaymentBottomSheet> {
  late final TextEditingController _amountController;
  late final TextEditingController _sessionsController;
  late StudentPaymentMode _paymentMode;
  DateTime? _paidFrom;
  DateTime? _paidUntil;

  @override
  void initState() {
    super.initState();
    _paymentMode = widget.currentCustomMode ?? widget.defaultPaymentMode;

    final amount = widget.currentCustomAmount ?? widget.defaultAmount;
    _amountController = TextEditingController(
      text: amount > 0 ? _formatAmount(amount) : '',
    );
    _amountController.addListener(_onChanged);

    final sessions = widget.currentCustomSessions ?? widget.defaultSessions;
    _sessionsController = TextEditingController(
      text: (sessions != null && sessions > 0) ? sessions.toString() : '',
    );
    _sessionsController.addListener(_onChanged);

    _paidFrom = widget.currentCustomPaidFrom ?? widget.defaultPaidFrom;
    _paidUntil = widget.currentCustomPaidUntil ?? widget.defaultPaidUntil;
  }

  void _onChanged() => setState(() {});

  @override
  void dispose() {
    _amountController.removeListener(_onChanged);
    _amountController.dispose();
    _sessionsController.removeListener(_onChanged);
    _sessionsController.dispose();
    super.dispose();
  }

  String _formatAmount(num amount) {
    final asDouble = amount.toDouble();
    return asDouble == asDouble.roundToDouble()
        ? asDouble.toInt().toString()
        : asDouble.toString();
  }

  DateTime get _maxPaidUntil =>
      DateTime.now().add(const Duration(days: 365 * 3));

  bool get _isDateRangeInvalid =>
      _paidFrom != null &&
      _paidUntil != null &&
      DateUtils.dateOnly(_paidFrom!).isAfter(DateUtils.dateOnly(_paidUntil!));

  void _onReset() {
    Navigator.of(context).pop(
      const GroupStudentPayment(
        selected: true,
        customPaymentMode: null,
        customAmount: null,
        customPaidFrom: null,
        customPaidUntil: null,
        customSessionsAdded: null,
      ),
    );
  }

  void _onSave() {
    final parsedAmount = double.tryParse(_amountController.text.trim());
    if (parsedAmount == null || parsedAmount <= 0) return;

    final isCustomMode = _paymentMode != widget.defaultPaymentMode;
    final isCustomAmount = parsedAmount != widget.defaultAmount;

    if (_paymentMode == StudentPaymentMode.paidUntil) {
      if (_paidUntil == null || _isDateRangeInvalid) return;

      final isCustomPaidFrom =
          (widget.defaultPaidFrom == null || _paidFrom == null)
          ? widget.defaultPaidFrom != _paidFrom
          : DateUtils.dateOnly(_paidFrom!) !=
                DateUtils.dateOnly(widget.defaultPaidFrom!);
      final isCustomPaidUntil =
          DateUtils.dateOnly(_paidUntil!) !=
          DateUtils.dateOnly(widget.defaultPaidUntil);

      final hasCustom =
          isCustomMode ||
          isCustomAmount ||
          isCustomPaidFrom ||
          isCustomPaidUntil;

      Navigator.of(context).pop(
        GroupStudentPayment(
          selected: true,
          customPaymentMode: hasCustom && isCustomMode ? _paymentMode : null,
          customAmount: hasCustom && isCustomAmount ? parsedAmount : null,
          customPaidFrom: hasCustom && isCustomPaidFrom ? _paidFrom : null,
          customPaidUntil: hasCustom && isCustomPaidUntil ? _paidUntil : null,
          customSessionsAdded: null,
        ),
      );
    } else {
      final parsedSessions = int.tryParse(_sessionsController.text.trim());
      if (parsedSessions == null || parsedSessions <= 0) return;

      final isCustomSessions = parsedSessions != widget.defaultSessions;
      final hasCustom = isCustomMode || isCustomAmount || isCustomSessions;

      Navigator.of(context).pop(
        GroupStudentPayment(
          selected: true,
          customPaymentMode: hasCustom && isCustomMode ? _paymentMode : null,
          customAmount: hasCustom && isCustomAmount ? parsedAmount : null,
          customPaidFrom: null,
          customPaidUntil: null,
          customSessionsAdded: hasCustom && isCustomSessions
              ? parsedSessions
              : null,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final keyboardInset = MediaQuery.of(context).viewInsets.bottom;
    final hasCustomValues =
        widget.currentCustomMode != null ||
        widget.currentCustomAmount != null ||
        widget.currentCustomPaidFrom != null ||
        widget.currentCustomPaidUntil != null ||
        widget.currentCustomSessions != null;

    final parsedAmount = double.tryParse(_amountController.text.trim());
    final isAmountValid = parsedAmount != null && parsedAmount > 0;
    final isDateValid = _paidUntil != null && !_isDateRangeInvalid;
    final parsedSessions = int.tryParse(_sessionsController.text.trim());
    final isSessionsValid = parsedSessions != null && parsedSessions > 0;
    final canSave =
        isAmountValid &&
        (_paymentMode == StudentPaymentMode.paidUntil
            ? isDateValid
            : isSessionsValid);

    return SafeArea(
      child: Padding(
        padding: EdgeInsets.only(
          left: 22,
          right: 22,
          bottom: 22 + keyboardInset,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          spacing: 16,
          children: [
            Text(
              widget.studentName,
              textAlign: TextAlign.center,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w600),
            ),
            StudentPaymentModeSelector(
              selectedMode: _paymentMode,
              onChanged: (mode) => setState(() => _paymentMode = mode),
            ),
            TextField(
              controller: _amountController,
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              inputFormatters: [
                FilteringTextInputFormatter.allow(RegExp(r'^\d*\.?\d{0,2}')),
              ],
              decoration: InputDecoration(
                labelText: '${context.t.finance.amount}*',
                hintText: '0.00',
                prefixIcon: const Icon(FluentIcons.money_24_regular),
              ),
            ),
            if (_paymentMode == StudentPaymentMode.paidUntil)
              Row(
                children: [
                  Expanded(
                    child: PaidFromField(
                      initialDate: _paidFrom,
                      lastDate: _maxPaidUntil,
                      enabled: true,
                      errorText: _isDateRangeInvalid
                          ? context.t.invalidDate
                          : null,
                      onDateChanged: (date) => setState(() => _paidFrom = date),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: PaidUntilField(
                      initialDate: _paidUntil,
                      lastDate: _maxPaidUntil,
                      enabled: true,
                      errorText: _isDateRangeInvalid
                          ? context.t.invalidDate
                          : null,
                      onDateChanged: (date) =>
                          setState(() => _paidUntil = date),
                    ),
                  ),
                ],
              )
            else
              SessionsCountField(
                controller: _sessionsController,
                enabled: true,
              ),
            const SizedBox(height: 8),
            Row(
              children: [
                if (hasCustomValues)
                  TextButton(
                    onPressed: _onReset,
                    child: Text(
                      context.t.reset,
                      style: const TextStyle(fontSize: 18),
                    ),
                  ),
                const Spacer(),
                TextButton(
                  onPressed: canSave ? _onSave : null,
                  child: Text(
                    context.t.create,
                    style: const TextStyle(fontSize: 18),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
