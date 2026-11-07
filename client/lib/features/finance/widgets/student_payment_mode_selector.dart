import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:phone/i18n/strings.g.dart';

enum StudentPaymentMode { paidUntil, sessionsCount }

class StudentPaymentModeSelector extends StatelessWidget {
  final StudentPaymentMode selectedMode;
  final ValueChanged<StudentPaymentMode> onChanged;

  const new({super.key, required this.selectedMode, required this.onChanged});

  @override
  Widget build(BuildContext context) => SizedBox(
    width: double.infinity,
    child: SegmentedButton<StudentPaymentMode>(
      showSelectedIcon: false,
      segments: [
        ButtonSegment<StudentPaymentMode>(
          value: StudentPaymentMode.paidUntil,
          label: Text(
            context.t.finance.paidUntil,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          icon: const Icon(FluentIcons.calendar_ltr_24_regular),
        ),
        ButtonSegment<StudentPaymentMode>(
          value: StudentPaymentMode.sessionsCount,
          label: Text(
            context.t.finance.sessionsCount,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          icon: const Icon(FluentIcons.ticket_diagonal_16_regular),
        ),
      ],
      selected: {selectedMode},
      onSelectionChanged: (selection) => onChanged(selection.first),
    ),
  );
}
