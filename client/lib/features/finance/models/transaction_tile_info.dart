import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:phone/core/constants/app_colors.dart';
import 'package:phone/core/constants/app_date_formats.dart';
import 'package:phone/generated/models/student_payment_details_response.dart';
import 'package:phone/generated/models/transaction_response.dart';
import 'package:phone/generated/models/transaction_type.dart';
import 'package:phone/i18n/strings.g.dart';

class TransactionTileInfo {
  final String title;
  final String subtitle;
  final IconData icon;
  final Color backgroundColor;

  const new({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.backgroundColor,
  });

  static TransactionTileInfo fromTransaction(
    TransactionResponse transaction, {
    String? overrideTitle,
  }) {
    final details = transaction.paymentDetails;
    final isIncome = transaction.type == TransactionType.income;
    final typeIcon = isIncome
        ? Icons.arrow_downward_rounded
        : Icons.arrow_upward_rounded;
    final typeColor = isIncome ? AppColors.income : AppColors.expense;

    if (details != null) {
      return TransactionTileInfo(
        title: overrideTitle ?? details.studentName ?? t.finance.deposit,
        subtitle: _subtitle(details),
        icon: typeIcon,
        backgroundColor: typeColor,
      );
    }

    return TransactionTileInfo(
      title: isIncome ? t.finance.deposit : t.finance.withdrawal,
      subtitle: '',
      icon: typeIcon,
      backgroundColor: typeColor,
    );
  }

  static String _subtitle(StudentPaymentDetailsResponse details) {
    final parts = <String>[
      if (details.paidUntil != null)
        '${t.finance.paidUntil} '
            '${DateFormat(AppDateFormats.dayMonthYear).format(details.paidUntil!)}',
      if (details.sessionsAdded != null)
        t.titleWithCount(
          title: t.finance.sessionsCount,
          count: details.sessionsAdded!,
        ),
    ];

    return parts.join(' • ');
  }
}
