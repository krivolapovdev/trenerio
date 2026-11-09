import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:phone/core/widgets/date_input_field.dart';
import 'package:phone/i18n/strings.g.dart';

class PaidFromField extends StatelessWidget {
  final DateTime? initialDate;
  final DateTime? firstDate;
  final DateTime? lastDate;
  final bool enabled;
  final bool isLoading;
  final String? errorText;
  final ValueChanged<DateTime?> onDateChanged;

  const new({
    super.key,
    required this.initialDate,
    this.firstDate,
    this.lastDate,
    this.isLoading = false,
    this.errorText,
    required this.enabled,
    required this.onDateChanged,
  });

  @override
  Widget build(BuildContext context) => DateInputField(
    labelText: context.t.finance.paidFrom,
    icon: FluentIcons.calendar_ltr_16_regular,
    initialDate: initialDate,
    firstDate: firstDate,
    lastDate: lastDate,
    enabled: enabled,
    allowClear: false,
    isLoading: isLoading,
    errorText: errorText,
    onDateChanged: onDateChanged,
  );
}
