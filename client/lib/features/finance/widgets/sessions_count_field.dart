import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:phone/i18n/strings.g.dart';

class SessionsCountField extends StatelessWidget {
  final TextEditingController controller;
  final bool enabled;
  final VoidCallback? onTap;

  const new({
    super.key,
    required this.controller,
    required this.enabled,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) => TextField(
    controller: controller,
    enabled: enabled,
    keyboardType: TextInputType.number,
    inputFormatters: [FilteringTextInputFormatter.digitsOnly],
    decoration: InputDecoration(
      labelText: '${context.t.finance.sessionsCount}*',
      prefixIcon: const Icon(FluentIcons.ticket_diagonal_16_regular),
    ),
    onTap: onTap,
  );
}
