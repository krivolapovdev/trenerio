import 'package:flutter/material.dart';
import 'package:phone/core/widgets/card_badge.dart';
import 'package:phone/features/students/models/student_badge_data.dart';

class StudentBadges extends StatelessWidget {
  static const double _spacing = 8.0;

  final List<StudentBadgeData> badges;
  final WrapAlignment alignment;

  const new({
    super.key,
    required this.badges,
    this.alignment = WrapAlignment.start,
  });

  @override
  Widget build(BuildContext context) => Wrap(
    alignment: alignment,
    spacing: _spacing,
    runSpacing: _spacing,
    children: [
      for (final badge in badges)
        CardBadge(
          key: ValueKey(badge.id),
          icon: badge.icon,
          label: badge.label,
          backgroundColor: badge.backgroundColor,
          foregroundColor: badge.foregroundColor,
        ),
    ],
  );
}
