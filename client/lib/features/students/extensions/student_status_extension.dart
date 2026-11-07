import 'package:flutter/material.dart';
import 'package:phone/core/constants/app_colors.dart';
import 'package:phone/features/students/models/student_badge_data.dart';
import 'package:phone/generated/models/student_status.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/i18n/strings.g.dart';

extension StudentStatusLabelExtension on StudentStatus {
  String label(BuildContext context) => switch (this) {
    StudentStatus.inactive => context.t.students.status.inactive,
    StudentStatus.present => context.t.students.status.present,
    StudentStatus.missing => context.t.students.status.missing,
    StudentStatus.paid => context.t.students.status.paid,
    StudentStatus.unpaid => context.t.students.status.unpaid,
    StudentStatus.free => context.t.students.status.free,
    StudentStatus.$unknown => '',
  };
}

extension StudentStatusExtension on List<StudentStatus> {
  List<StudentBadgeData> toBadges(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

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

        StudentStatus.paid => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: colorScheme.primaryContainer.withValues(alpha: 0.5),
          foregroundColor: colorScheme.primary,
        ),

        StudentStatus.unpaid => StudentBadgeData(
          id: status.name,
          label: status.label(context),
          backgroundColor: colorScheme.errorContainer,
          foregroundColor: colorScheme.onErrorContainer,
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
  /// - present / missing
  /// - paid / unpaid / free (or inactive)
  List<StudentBadgeData> toHeaderStatusBadges(BuildContext context) {
    final statusSet = statuses.toSet();

    // 1. Attendance status: present / missing
    final StudentStatus? attendance;
    if (statusSet.contains(StudentStatus.present)) {
      attendance = StudentStatus.present;
    } else if (statusSet.contains(StudentStatus.missing)) {
      attendance = StudentStatus.missing;
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
