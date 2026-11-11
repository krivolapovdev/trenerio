import 'package:flutter/material.dart';
import 'package:phone/core/constants/app_colors.dart';
import 'package:phone/features/students/models/student_badge_data.dart';
import 'package:phone/generated/models/student_status.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/i18n/strings.g.dart';

extension StudentStatusLabelExtension on StudentStatus {
  String label([BuildContext? context]) => switch (this) {
    StudentStatus.inactive =>
      context?.t.students.status.inactive ?? t.students.status.inactive,
    StudentStatus.present =>
      context?.t.students.status.present ?? t.students.status.present,
    StudentStatus.missing =>
      context?.t.students.status.missing ?? t.students.status.missing,
    StudentStatus.excused =>
      context?.t.students.status.excused ?? t.students.status.excused,
    StudentStatus.paid =>
      context?.t.students.status.paid ?? t.students.status.paid,
    StudentStatus.unpaid =>
      context?.t.students.status.unpaid ?? t.students.status.unpaid,
    StudentStatus.free =>
      context?.t.students.status.free ?? t.students.status.free,
    StudentStatus.$unknown => '',
  };
}

extension StudentStatusExtension on List<StudentStatus> {
  List<StudentBadgeData> toBadges([BuildContext? context]) {
    final colorScheme = context != null ? Theme.of(context).colorScheme : null;

    return map(
      (status) => switch (status) {
        StudentStatus.inactive => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: AppColors.neutralSurface,
          foregroundColor: AppColors.neutralOn,
        ),

        StudentStatus.present => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: AppColors.successSurface,
          foregroundColor: AppColors.successOn,
        ),

        StudentStatus.missing => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: AppColors.warningSurface.withValues(alpha: 0.8),
          foregroundColor: AppColors.warningOn,
        ),

        StudentStatus.excused => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: AppColors.neutralSurface,
          foregroundColor: AppColors.neutralOn,
        ),

        StudentStatus.paid => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: colorScheme != null
              ? colorScheme.primaryContainer.withValues(alpha: 0.5)
              : const Color(0xFFE8DEF8),
          foregroundColor: colorScheme?.primary ?? const Color(0xFF6750A4),
        ),

        StudentStatus.unpaid => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor:
              colorScheme?.errorContainer ?? const Color(0xFFFFD8D8),
          foregroundColor:
              colorScheme?.onErrorContainer ?? const Color(0xFFBA1A1A),
        ),

        StudentStatus.free => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: AppColors.freeSurface,
          foregroundColor: AppColors.freeOn,
        ),

        StudentStatus.$unknown => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: Colors.transparent,
          foregroundColor: Colors.transparent,
        ),
      },
    ).toList();
  }
}

extension StudentSummaryHeaderBadgesExtension on StudentSummaryResponse {
  /// First line header badges:
  /// - present / missing / excused
  /// - paid / unpaid / free (or inactive)
  List<StudentBadgeData> toHeaderStatusBadges([BuildContext? context]) {
    final statusSet = statuses.toSet();

    // 1. Attendance status: present / missing / excused
    final StudentStatus? attendance;
    if (statusSet.contains(StudentStatus.present)) {
      attendance = StudentStatus.present;
    } else if (statusSet.contains(StudentStatus.missing)) {
      attendance = StudentStatus.missing;
    } else if (statusSet.contains(StudentStatus.excused)) {
      attendance = StudentStatus.excused;
    } else {
      attendance = null;
    }

    // 2. Payment status: paid / unpaid / free (or inactive)
    final StudentStatus? payment;
    if (free || statusSet.contains(StudentStatus.free)) {
      payment = StudentStatus.free;
    } else if (statusSet.contains(StudentStatus.paid)) {
      payment = StudentStatus.paid;
    } else if (statusSet.contains(StudentStatus.unpaid)) {
      payment = StudentStatus.unpaid;
    } else if (statusSet.contains(StudentStatus.inactive)) {
      payment = StudentStatus.inactive;
    } else {
      payment = null;
    }

    final headerStatuses = <StudentStatus>[?attendance, ?payment];

    return headerStatuses.toBadges(context);
  }
}
